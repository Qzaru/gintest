package products

import (
	"context"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"go.uber.org/zap"
)

func NewProductsRedisStore() *RedisStore {
	return &RedisStore{
		Expiration: time.Second * 3600,
		//PreKey:     "P",
		Context: context.TODO(),
	}
}

type RedisStore struct {
	Expiration time.Duration
	//PreKey     string
	Context context.Context
}

func (rs *RedisStore) UseWithCtx(ctx context.Context) *RedisStore {
	if ctx == nil {
		rs.Context = ctx
	}
	return rs
}

func (rs *RedisStore) Set(skuid string, value string) error {
	err := global.GVA_REDIS.Set(rs.Context, skuid, value, rs.Expiration).Err()
	if err != nil {
		global.GVA_LOG.Error("商品信息存入缓存失败", zap.Error(err))
		return err
	}
	return nil
}

func (rs *RedisStore) Get(key string, clear bool) string {
	val, err := global.GVA_REDIS.Get(rs.Context, key).Result()
	if err != nil {
		global.GVA_LOG.Error("商品信息缓存获取失败", zap.Error(err))
		return ""
	}
	if clear {
		err := global.GVA_REDIS.Del(rs.Context, key).Err()
		if err != nil {
			global.GVA_LOG.Error("商品信息缓存清理失败", zap.Error(err))
			return ""
		}
	}
	return val
}

func (rs *RedisStore) Verify(skuid, answer string, clear bool) bool {
	key := skuid
	v := rs.Get(key, clear)
	return v == answer
}
