package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/solobueno/erp/internal/shared/observability"
)

// accessLogFields is a shared mutable holder for fields discovered while a
// request is being handled (e.g. user_id/tenant_id resolved by RequireAuth,
// which runs deeper in the chain with its own r.WithContext-derived request
// and so can't hand a new context value back up to AccessLog). AccessLog
// stashes a pointer to one of these in the context before calling next, and
// reads it after next returns.
type accessLogFields struct {
	UserID   string
	TenantID string
}

type accessLogContextKey struct{}

func setAccessLogUserID(ctx context.Context, userID string) {
	if f, ok := ctx.Value(accessLogContextKey{}).(*accessLogFields); ok {
		f.UserID = userID
	}
}

func setAccessLogTenantID(ctx context.Context, tenantID string) {
	if f, ok := ctx.Value(accessLogContextKey{}).(*accessLogFields); ok {
		f.TenantID = tenantID
	}
}

// AccessLog logs one structured "request completed" entry per request,
// regardless of outcome - including expected auth rejections (revoked/
// expired token, wrong password, insufficient role) that previously
// produced no structured log at all, only chi's bare unstructured line.
// Never logs email, password, or token values (Constitution XII).
func AccessLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fields := &accessLogFields{}
		ctx := context.WithValue(r.Context(), accessLogContextKey{}, fields)
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		start := time.Now()
		next.ServeHTTP(ww, r.WithContext(ctx))
		duration := time.Since(start)

		logFields := []observability.Field{
			{Key: "request_id", Value: middleware.GetReqID(ctx)},
			{Key: "method", Value: r.Method},
			{Key: "path", Value: r.URL.Path},
			{Key: "status", Value: ww.Status()},
			{Key: "duration_ms", Value: duration.Milliseconds()},
		}
		if fields.UserID != "" {
			logFields = append(logFields, observability.Field{Key: "user_id", Value: fields.UserID})
		}
		if fields.TenantID != "" {
			logFields = append(logFields, observability.Field{Key: "tenant_id", Value: fields.TenantID})
		}

		if ww.Status() >= 500 {
			logger.Error("request completed", logFields...)
		} else {
			logger.Info("request completed", logFields...)
		}
	})
}
