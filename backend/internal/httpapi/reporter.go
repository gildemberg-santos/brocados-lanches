package httpapi

import (
	"context"

	"github.com/getsentry/sentry-go"
)

// Reporter publica erros inesperados sem expor detalhes da infraestrutura às
// camadas superiores. A implementação neutra mantém os testes sem rede.
type Reporter interface {
	Capture(ctx context.Context, err error, tags map[string]string)
	Recover(ctx context.Context, value any, tags map[string]string)
}

// SentryReporter usa o hub da requisição, criado pelo middleware de escopo.
type SentryReporter struct{}

func (SentryReporter) Capture(ctx context.Context, err error, tags map[string]string) {
	hub := hubForReporting(ctx)
	if hub == nil {
		return
	}
	hub.WithScope(func(scope *sentry.Scope) {
		setTags(scope, tags)
		hub.CaptureException(err)
	})
}

func (SentryReporter) Recover(ctx context.Context, value any, tags map[string]string) {
	hub := hubForReporting(ctx)
	if hub == nil {
		return
	}
	hub.WithScope(func(scope *sentry.Scope) {
		setTags(scope, tags)
		hub.RecoverWithContext(ctx, value)
	})
}

func hubForReporting(ctx context.Context) *sentry.Hub {
	hub := sentry.GetHubFromContext(ctx)
	if hub == nil {
		hub = sentry.CurrentHub()
	}
	if hub == nil || hub.Client() == nil {
		return nil
	}
	return hub
}

func setTags(scope *sentry.Scope, tags map[string]string) {
	for key, value := range tags {
		scope.SetTag(key, value)
	}
}

// noopReporter é usado quando o DSN não está configurado.
type noopReporter struct{}

func (noopReporter) Capture(context.Context, error, map[string]string) {}

func (noopReporter) Recover(context.Context, any, map[string]string) {}

type reporterContextKey struct{}

func withReporter(ctx context.Context, reporter Reporter) context.Context {
	return context.WithValue(ctx, reporterContextKey{}, reporter)
}

func reporterFrom(ctx context.Context) Reporter {
	if reporter, ok := ctx.Value(reporterContextKey{}).(Reporter); ok {
		return reporter
	}
	return noopReporter{}
}
