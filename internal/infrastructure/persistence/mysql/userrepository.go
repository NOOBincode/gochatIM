package mysql

import (
	"context"
	"errors"
	"fmt"
	"time"

	"GochatIM/internal/domain/entity"
	"GochatIM/internal/domain/event"
	"GochatIM/internal/domain/repo/user"
	"GochatIM/internal/infrastructure/cache"
	"GochatIM/internal/infrastructure/messaging"
	"GochatIM/internal/infrastructure/task"
	"GochatIM/pkg/logger"
	"GochatIM/pkg/singleflight"
	"gorm.io/gorm"
)

// 定义Kafka主题
const (
	UserEventTopic = "user_events"
)

// UserModel GORM用户模型
type UserModel struct {
	ID        uint64 `gorm:"primaryKey"`
	Username  string `gorm:"uniqueIndex;size:50;not null"`
	Nickname  string `gorm:"size:50"`
	Password  string `gorm:"size:100;not null"`
	Email     string `gorm:"size:100;uniqueIndex"`
	Phone     string `gorm:"size:20;uniqueIndex"`
	Avatar    string `gorm:"size:255"`
	Status    int8   `gorm:"default:1"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// TableName 设置表名
func (UserModel) TableName() string {
	return "users"
}

// UserRepositoryImpl 用户仓储GORM实现
type UserRepositoryImpl struct {
	db            *gorm.DB
	userCache     cache.UserCache
	kafkaProducer *messaging.KafkaProducer
	sf            *singleflight.Group
	taskClient    *task.Client
}

// NewUserRepository 创建用户仓储
func NewUserRepository(db *gorm.DB, userCache cache.UserCache, kafkaProducer *messaging.KafkaProducer, sf *singleflight.Group, taskClient *task.Client) user.Repository {
	return &UserRepositoryImpl{
		db:            db,
		userCache:     userCache,
		kafkaProducer: kafkaProducer,
		sf:            sf,
		taskClient:    taskClient,
	}
}

// 将实体转换为模型
func entityToModel(user *entity.User) *UserModel {
	return &UserModel{
		ID:        user.ID,
		Username:  user.Username,
		Nickname:  user.Nickname,
		Password:  user.Password,
		Email:     user.Email,
		Phone:     user.Phone,
		Avatar:    user.Avatar,
		Status:    user.Status,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

// 将模型转换为实体
func modelToEntity(model *UserModel) *entity.User {
	return &entity.User{
		ID:        model.ID,
		Username:  model.Username,
		Nickname:  model.Nickname,
		Password:  model.Password,
		Email:     model.Email,
		Phone:     model.Phone,
		Avatar:    model.Avatar,
		Status:    model.Status,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}
}

// FindByID 根据ID查找用户
func (r *UserRepositoryImpl) FindByID(ctx context.Context, id uint64) (*entity.User, error) {
	// 先查缓存
	user, err := r.userCache.GetUser(ctx, id)
	if err == nil && user != nil {
		logger.Debugf("用户缓存命中: userID=%d", id)
		return user, nil
	}

	key := fmt.Sprintf("user:%d", id)
	v, err, _ := r.sf.Do(key, func() (interface{}, error) {
		user, err := r.userCache.GetUser(ctx, id)
		if err == nil && user != nil {
			logger.Infof("用户缓存命中(singleflight二次检查):userID=%d", id)
			return user, nil
		}
		//查数据库
		var model UserModel
		if err := r.db.WithContext(ctx).First(&model, id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("用户不存在")
			}
			logger.Errorf("查询用户失败: %v", err)
			return nil, err
		}

		// 转换为实体
		user = modelToEntity(&model)

		// 更新缓存
		if err := r.userCache.SetUser(ctx, user); err != nil {
			logger.Warnf("更新用户缓存失败: %v", err)
		}

		return user, nil
	})
	if err != nil {
		return nil, err
	}
	return v.(*entity.User), nil
}

// FindByUsername 根据用户名查找用户
func (r *UserRepositoryImpl) FindByUsername(ctx context.Context, username string) (*entity.User, error) {
	key := fmt.Sprintf("username:%s", username)
	v, err, _ := r.sf.Do(key, func() (interface{}, error) {
		var model UserModel
		if err := r.db.WithContext(ctx).Where("username = ?", username).First(&model).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("用户不存在")
			}
			logger.Errorf("查询用户失败: %v", err)
			return nil, err
		}

		// 转换为实体
		user := modelToEntity(&model)

		// 更新缓存
		if err := r.userCache.SetUser(ctx, user); err != nil {
			logger.Warnf("更新用户缓存失败: %v", err)
		}

		return user, nil
	})

	if err != nil {
		return nil, err
	}
	return v.(*entity.User), nil
}

// Create 创建用户
func (r *UserRepositoryImpl) Create(ctx context.Context, user *entity.User) error {
	// 转换为模型
	model := entityToModel(user)

	// 保存到数据库
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		logger.Errorf("创建用户失败: %v", err)
		return err
	}

	// 更新ID
	user.ID = model.ID

	// 更新缓存
	if err := r.userCache.SetUser(ctx, user); err != nil {
		logger.Warnf("更新用户缓存失败: %v", err)
	}

	// 发送Kafka事件
	userEvent := event.NewUserEvent(event.UserCreated, user.ID, user.Username)
	if err := r.kafkaProducer.SendMessage(UserEventTopic, user.Username, userEvent); err != nil {
		logger.Warnf("发送用户创建事件失败: %v", err)
	}

	return nil
}

// Update 更新用户
func (r *UserRepositoryImpl) Update(ctx context.Context, user *entity.User) error {
	// 转换为模型
	model := entityToModel(user)

	// 更新数据库
	if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
		logger.Errorf("更新用户失败: %v", err)
		return err
	}

	// 更新缓存
	if err := r.userCache.SetUser(ctx, user); err != nil {
		logger.Warnf("更新用户缓存失败: %v", err)
	}

	// 发送Kafka事件
	userEvent := event.NewUserEvent(event.UserUpdated, user.ID, user.Username)
	if err := r.kafkaProducer.SendMessage(UserEventTopic, user.Username, userEvent); err != nil {
		logger.Warnf("发送用户更新事件失败: %v", err)
	}

	return nil
}

// Delete 删除用户
func (r *UserRepositoryImpl) Delete(ctx context.Context, id uint64) error {
	// 先查询用户
	var model UserModel
	if err := r.db.WithContext(ctx).First(&model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("用户不存在")
		}
		logger.Errorf("查询用户失败: %v", err)
		return err
	}

	// 删除数据库记录
	if err := r.db.WithContext(ctx).Delete(&model).Error; err != nil {
		logger.Errorf("删除用户失败: %v", err)
		return err
	}

	// 删除缓存
	if err := r.userCache.DeleteUser(ctx, id); err != nil {
		logger.Warnf("删除用户缓存失败: %v", err)
	}

	// 发送Kafka事件
	userEvent := event.NewUserEvent(event.UserDeleted, id, model.Username)
	if err := r.kafkaProducer.SendMessage(UserEventTopic, model.Username, userEvent); err != nil {
		logger.Warnf("发送用户删除事件失败: %v", err)
	}

	return nil
}

// FindByIDs 批量查询用户
func (r *UserRepositoryImpl) FindByIDs(ctx context.Context, ids []uint64) ([]*entity.User, error) {
	var models []UserModel
	if err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&models).Error; err != nil {
		logger.Errorf("批量查询用户失败: %v", err)
		return nil, err
	}

	users := make([]*entity.User, len(models))
	for i, model := range models {
		users[i] = modelToEntity(&model)

		// 更新缓存
		if err := r.userCache.SetUser(ctx, users[i]); err != nil {
			logger.Warnf("更新用户缓存失败: %v", err)
		}
	}

	return users, nil
}

// FindByEmail 根据邮箱查找用户
func (r *UserRepositoryImpl) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	var model UserModel
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		logger.Errorf("查询用户失败: %v", err)
		return nil, err
	}

	// 转换为实体
	user := modelToEntity(&model)

	// 更新缓存
	if err := r.userCache.SetUser(ctx, user); err != nil {
		logger.Warnf("更新用户缓存失败: %v", err)
	}

	return user, nil
}

// FindByPhone 根据手机号查找用户
func (r *UserRepositoryImpl) FindByPhone(ctx context.Context, phone string) (*entity.User, error) {
	var model UserModel
	if err := r.db.WithContext(ctx).Where("phone = ?", phone).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		logger.Errorf("查询用户失败: %v", err)
		return nil, err
	}

	// 转换为实体
	user := modelToEntity(&model)

	// 更新缓存
	if err := r.userCache.SetUser(ctx, user); err != nil {
		logger.Warnf("更新用户缓存失败: %v", err)
	}

	return user, nil
}
