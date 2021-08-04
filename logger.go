package main

import (
	"context"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

// ContextKey is used for context.Context value. The value requires a key that is not primitive type.
type ContextKey string // can be unexported

// ContextKeyRequestID is the ContextKey for RequestID
const ContextKeyRequestID ContextKey = "requestId" // can be unexported

// AssignRequestID will attach a brand new request ID to a http request
func AssignRequestID(ctx context.Context) context.Context {
	reqID := uuid.New()
	return context.WithValue(ctx, ContextKeyRequestID, reqID.String())
}

// GetRequestID will get reqID from a http request and return it as a string
func GetRequestID(ctx context.Context) string {
	reqID := ctx.Value(ContextKeyRequestID)
	if ret, ok := reqID.(string); ok {
		return ret
	}
	return ""
}

func GetRequestLogger(r *http.Request) *log.Entry {
	rid := GetRequestID(r.Context())
	return log.WithField(string(ContextKeyRequestID), rid)
}

//#region middlewares
func reqIDMiddleware1(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		r = r.WithContext(AssignRequestID(ctx))
		entry := log.WithField("requestId", GetRequestID(r.Context()))
		entry.Infof("Incomming request %s %s %s", r.Method, r.RequestURI, r.RemoteAddr)
		timeStart := time.Now()
		next(w, r)
		entry.Infof("Finished handling http req (%dms)", time.Since(timeStart).Milliseconds())
	}
}
