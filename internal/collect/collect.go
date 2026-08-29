package collect

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/hwchiu/SalaryThief/internal/config"
	"github.com/hwchiu/SalaryThief/internal/model"
)

func ErrorClass(err error) model.ErrorClass {
	if err == nil {
		return model.ErrorNone
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return model.ErrorTimeout
	}
	var uerr *url.Error
	if errors.As(err, &uerr) {
		if errors.Is(uerr.Err, context.DeadlineExceeded) {
			return model.ErrorTimeout
		}
		var tlsErr tls.RecordHeaderError
		if errors.As(uerr.Err, &tlsErr) {
			return model.ErrorTLS
		}
	}
	var ne net.Error
	if errors.As(err, &ne) {
		if ne.Timeout() {
			return model.ErrorTimeout
		}
		return model.ErrorNetwork
	}
	var he *httpError
	if errors.As(err, &he) {
		if he.code == http.StatusUnauthorized || he.code == http.StatusForbidden {
			return model.ErrorAuth
		}
		return model.ErrorHTTP
	}
	return model.ErrorUnknown
}

type httpError struct{ code int }

func (e *httpError) Error() string { return fmt.Sprintf("unexpected Redfish HTTP status %d", e.code) }

// Scrape makes all Redfish requests. It is deliberately separate from metrics serving.
func Scrape(ctx context.Context, t config.Target, _ *slog.Logger) model.Snapshot {
	started := time.Now()
	s := model.Snapshot{Target: t.Name, Scope: t.Labels["observability_scope"], Labels: t.Labels, LastAttempt: started, Resources: map[string]model.ResourceStatus{}}
	if s.Scope == "" {
		s.Scope = "default"
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: t.InsecureSkipVerify} // #nosec G402: explicit local/config choice
	client := &http.Client{Transport: transport, Timeout: t.Timeout}
	if err := get(ctx, client, t, "/redfish/v1"); err != nil {
		s.ErrorClass = ErrorClass(err)
		s.Duration = time.Since(started)
		return s
	}
	s.Up = true
	s.LastSuccess = time.Now()
	resources := []struct {
		name, path string
		enabled    bool
	}{
		{"system", "/redfish/v1/Systems/1", true}, {"thermal", "/redfish/v1/Chassis/1/Thermal", t.Collect.Thermal}, {"power", "/redfish/v1/Chassis/1/Power", t.Collect.Power}, {"storage", "/redfish/v1/Systems/1/Storage", t.Collect.Storage},
	}
	for _, r := range resources {
		if !r.enabled {
			continue
		}
		status := model.ResourceStatus{State: model.HealthOK, LastAttempt: time.Now()}
		err := get(ctx, client, t, r.path)
		if err != nil {
			status.State = model.HealthError
			status.ErrorClass = ErrorClass(err)
			if s.ErrorClass == model.ErrorNone {
				s.ErrorClass = model.ErrorPartial
			}
		} else {
			status.LastSuccess = time.Now()
		}
		s.Resources[r.name] = status
	}
	s.Duration = time.Since(started)
	return s
}
func get(ctx context.Context, client *http.Client, target config.Target, path string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(target.Endpoint, "/")+path, nil)
	if err != nil {
		return err
	}
	if target.Username != "" {
		req.SetBasicAuth(target.Username, target.Password)
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return &httpError{resp.StatusCode}
	}
	return nil
}
