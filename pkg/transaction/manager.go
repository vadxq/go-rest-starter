package transaction

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// Manager 事务管理器接口
type Manager interface {
	// Execute 执行事务
	Execute(ctx context.Context, fn TxFunc) error
	// ExecuteWithOptions 使用选项执行事务
	ExecuteWithOptions(ctx context.Context, opts *sql.TxOptions, fn TxFunc) error
}

// TxFunc 事务函数类型
type TxFunc func(ctx context.Context, tx *gorm.DB) error

// GormTransactionManager GORM事务管理器
type GormTransactionManager struct {
	db *gorm.DB
}

// NewGormTransactionManager 创建GORM事务管理器
func NewGormTransactionManager(db *gorm.DB) Manager {
	return &GormTransactionManager{db: db}
}

// Execute 执行事务
func (m *GormTransactionManager) Execute(ctx context.Context, fn TxFunc) error {
	return m.ExecuteWithOptions(ctx, nil, fn)
}

// ExecuteWithOptions 使用选项执行事务
func (m *GormTransactionManager) ExecuteWithOptions(ctx context.Context, opts *sql.TxOptions, fn TxFunc) error {
	// 开始事务
	tx := m.db.WithContext(ctx)
	if opts != nil {
		tx = tx.Begin(opts)
	} else {
		tx = tx.Begin()
	}
	
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	// 使用defer确保事务一定会结束
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r) // 重新抛出panic
		}
	}()

	// 执行事务函数
	if err := fn(ctx, tx); err != nil {
		// 回滚事务
		if rbErr := tx.Rollback().Error; rbErr != nil {
			return fmt.Errorf("failed to rollback transaction: %v (original error: %w)", rbErr, err)
		}
		return err
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// RunInTransaction 在事务中运行函数（简化版）
func RunInTransaction(db *gorm.DB, fn func(*gorm.DB) error) error {
	tx := db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// Transactional 事务装饰器
type Transactional struct {
	db *gorm.DB
}

// NewTransactional 创建事务装饰器
func NewTransactional(db *gorm.DB) *Transactional {
	return &Transactional{db: db}
}

// Wrap 包装函数使其在事务中执行
func (t *Transactional) Wrap(fn func(*gorm.DB) error) error {
	return RunInTransaction(t.db, fn)
}

// NestedTransaction 嵌套事务支持
type NestedTransaction struct {
	db           *gorm.DB
	savepoints   []string
	currentLevel int
}

// NewNestedTransaction 创建嵌套事务
func NewNestedTransaction(db *gorm.DB) *NestedTransaction {
	return &NestedTransaction{
		db:           db,
		savepoints:   make([]string, 0),
		currentLevel: 0,
	}
}

// Begin 开始新的事务或保存点
func (nt *NestedTransaction) Begin() error {
	if nt.currentLevel == 0 {
		// 开始新事务
		tx := nt.db.Begin()
		if tx.Error != nil {
			return tx.Error
		}
		nt.db = tx
	} else {
		// 创建保存点
		savepoint := fmt.Sprintf("sp_%d", nt.currentLevel)
		if err := nt.db.Exec("SAVEPOINT " + savepoint).Error; err != nil {
			return err
		}
		nt.savepoints = append(nt.savepoints, savepoint)
	}
	nt.currentLevel++
	return nil
}

// Commit 提交事务或释放保存点
func (nt *NestedTransaction) Commit() error {
	if nt.currentLevel == 0 {
		return errors.New("no transaction to commit")
	}

	if nt.currentLevel == 1 {
		// 提交事务
		if err := nt.db.Commit().Error; err != nil {
			return err
		}
	} else {
		// 释放保存点
		savepoint := nt.savepoints[len(nt.savepoints)-1]
		if err := nt.db.Exec("RELEASE SAVEPOINT " + savepoint).Error; err != nil {
			return err
		}
		nt.savepoints = nt.savepoints[:len(nt.savepoints)-1]
	}
	nt.currentLevel--
	return nil
}

// Rollback 回滚事务或回滚到保存点
func (nt *NestedTransaction) Rollback() error {
	if nt.currentLevel == 0 {
		return errors.New("no transaction to rollback")
	}

	if nt.currentLevel == 1 {
		// 回滚整个事务
		if err := nt.db.Rollback().Error; err != nil {
			return err
		}
	} else {
		// 回滚到保存点
		savepoint := nt.savepoints[len(nt.savepoints)-1]
		if err := nt.db.Exec("ROLLBACK TO SAVEPOINT " + savepoint).Error; err != nil {
			return err
		}
		nt.savepoints = nt.savepoints[:len(nt.savepoints)-1]
	}
	nt.currentLevel--
	return nil
}

// TransactionContext 事务上下文
type TransactionContext struct {
	ctx context.Context
	tx  *gorm.DB
}

// NewTransactionContext 创建事务上下文
func NewTransactionContext(ctx context.Context, db *gorm.DB) (*TransactionContext, error) {
	tx := db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	
	return &TransactionContext{
		ctx: ctx,
		tx:  tx,
	}, nil
}

// DB 获取事务数据库连接
func (tc *TransactionContext) DB() *gorm.DB {
	return tc.tx
}

// Context 获取上下文
func (tc *TransactionContext) Context() context.Context {
	return tc.ctx
}

// Commit 提交事务
func (tc *TransactionContext) Commit() error {
	return tc.tx.Commit().Error
}

// Rollback 回滚事务
func (tc *TransactionContext) Rollback() error {
	return tc.tx.Rollback().Error
}

// Complete 根据error决定提交或回滚
func (tc *TransactionContext) Complete(err error) error {
	if err != nil {
		if rbErr := tc.Rollback(); rbErr != nil {
			return fmt.Errorf("rollback failed: %v, original error: %w", rbErr, err)
		}
		return err
	}
	return tc.Commit()
}

// WithTransaction 在事务中执行函数（带上下文）
func WithTransaction(ctx context.Context, db *gorm.DB, fn func(context.Context, *gorm.DB) error) error {
	tc, err := NewTransactionContext(ctx, db)
	if err != nil {
		return err
	}
	
	defer func() {
		if r := recover(); r != nil {
			tc.Rollback()
			panic(r)
		}
	}()
	
	err = fn(tc.Context(), tc.DB())
	return tc.Complete(err)
}