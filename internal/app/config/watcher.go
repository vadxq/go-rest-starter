package config

import (
	"log/slog"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// ConfigWatcher 配置文件监听器
type ConfigWatcher struct {
	mu        sync.RWMutex
	config    *AppConfig
	callbacks []func(*AppConfig)
	watcher   *fsnotify.Watcher
	stopCh    chan struct{}
}

// NewConfigWatcher 创建配置监听器
func NewConfigWatcher(configPath string) (*ConfigWatcher, error) {
	// 初始加载配置
	cfg, err := LoadConfig(configPath)
	if err != nil {
		return nil, err
	}

	// 创建文件监听器
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	// 添加配置文件到监听
	if err := watcher.Add(configPath); err != nil {
		watcher.Close()
		return nil, err
	}

	cw := &ConfigWatcher{
		config:    cfg,
		callbacks: make([]func(*AppConfig), 0),
		watcher:   watcher,
		stopCh:    make(chan struct{}),
	}

	// 启动监听
	go cw.watch(configPath)

	return cw, nil
}

// watch 监听配置文件变化
func (cw *ConfigWatcher) watch(configPath string) {
	// 防抖定时器
	var debounceTimer *time.Timer
	debounceDuration := 100 * time.Millisecond

	for {
		select {
		case event, ok := <-cw.watcher.Events:
			if !ok {
				return
			}

			// 只处理写入和创建事件
			if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
				// 使用防抖处理，避免频繁重载
				if debounceTimer != nil {
					debounceTimer.Stop()
				}
				
				debounceTimer = time.AfterFunc(debounceDuration, func() {
					cw.reloadConfig(configPath)
				})
			}

		case err, ok := <-cw.watcher.Errors:
			if !ok {
				return
			}
			slog.Error("配置文件监听错误", "error", err)

		case <-cw.stopCh:
			if debounceTimer != nil {
				debounceTimer.Stop()
			}
			return
		}
	}
}

// reloadConfig 重新加载配置
func (cw *ConfigWatcher) reloadConfig(configPath string) {
	slog.Info("检测到配置文件变化，重新加载配置", "path", configPath)

	// 重新读取配置
	newCfg, err := LoadConfig(configPath)
	if err != nil {
		slog.Error("重新加载配置失败", "error", err)
		return
	}

	// 验证新配置
	if err := cw.validateConfig(newCfg); err != nil {
		slog.Error("配置验证失败", "error", err)
		return
	}

	// 更新配置
	cw.mu.Lock()
	oldCfg := cw.config
	cw.config = newCfg
	cw.mu.Unlock()

	slog.Info("配置重新加载成功")

	// 通知所有回调
	cw.notifyCallbacks(oldCfg, newCfg)
}

// validateConfig 验证配置
func (cw *ConfigWatcher) validateConfig(cfg *AppConfig) error {
	// 这里可以添加配置验证逻辑
	// 例如：检查必填字段、验证格式等
	
	if cfg.Server.Port <= 0 || cfg.Server.Port > 65535 {
		return ErrInvalidPort
	}

	if cfg.Database.Host == "" {
		return ErrMissingDatabaseHost
	}

	return nil
}

// notifyCallbacks 通知所有回调函数
func (cw *ConfigWatcher) notifyCallbacks(oldCfg, newCfg *AppConfig) {
	// 记录配置变化
	cw.logConfigChanges(oldCfg, newCfg)

	// 执行回调
	for _, callback := range cw.callbacks {
		go func(cb func(*AppConfig)) {
			defer func() {
				if r := recover(); r != nil {
					slog.Error("配置变更回调执行失败", "error", r)
				}
			}()
			cb(newCfg)
		}(callback)
	}
}

// logConfigChanges 记录配置变化
func (cw *ConfigWatcher) logConfigChanges(oldCfg, newCfg *AppConfig) {
	// 记录主要配置变化
	if oldCfg.Server.Port != newCfg.Server.Port {
		slog.Info("服务端口变更", "old", oldCfg.Server.Port, "new", newCfg.Server.Port)
	}
	
	if oldCfg.Log.Level != newCfg.Log.Level {
		slog.Info("日志级别变更", "old", oldCfg.Log.Level, "new", newCfg.Log.Level)
	}
	
	if oldCfg.Database.MaxOpenConns != newCfg.Database.MaxOpenConns {
		slog.Info("数据库最大连接数变更", "old", oldCfg.Database.MaxOpenConns, "new", newCfg.Database.MaxOpenConns)
	}
}

// GetConfig 获取当前配置
func (cw *ConfigWatcher) GetConfig() *AppConfig {
	cw.mu.RLock()
	defer cw.mu.RUnlock()
	return cw.config
}

// OnConfigChange 注册配置变更回调
func (cw *ConfigWatcher) OnConfigChange(callback func(*AppConfig)) {
	cw.mu.Lock()
	defer cw.mu.Unlock()
	cw.callbacks = append(cw.callbacks, callback)
}

// Stop 停止监听
func (cw *ConfigWatcher) Stop() error {
	close(cw.stopCh)
	return cw.watcher.Close()
}

// WatchConfig 监听配置文件变化（使用Viper）
func WatchConfig(onChange func()) {
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		slog.Info("配置文件已更改", "file", e.Name)
		if onChange != nil {
			onChange()
		}
	})
}

// HotReloadConfig 热重载配置示例
type HotReloadConfig struct {
	mu     sync.RWMutex
	config *AppConfig
}

// NewHotReloadConfig 创建热重载配置
func NewHotReloadConfig(configPath string) (*HotReloadConfig, error) {
	cfg, err := LoadConfig(configPath)
	if err != nil {
		return nil, err
	}

	hrc := &HotReloadConfig{
		config: cfg,
	}

	// 设置Viper监听
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		slog.Info("配置文件变化，重新加载", "file", e.Name)
		
		// 重新解析配置
		newCfg := &AppConfig{}
		if err := viper.Unmarshal(newCfg); err != nil {
			slog.Error("解析新配置失败", "error", err)
			return
		}

		// 更新配置
		hrc.mu.Lock()
		hrc.config = newCfg
		hrc.mu.Unlock()
		
		slog.Info("配置热重载成功")
	})

	return hrc, nil
}

// Get 获取配置（线程安全）
func (hrc *HotReloadConfig) Get() *AppConfig {
	hrc.mu.RLock()
	defer hrc.mu.RUnlock()
	return hrc.config
}