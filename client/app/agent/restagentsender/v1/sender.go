package v1

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"time"

	"github.com/AlexBlackNn/metrics/app/agent/encryption"
	"github.com/AlexBlackNn/metrics/app/agent/hash"
	"github.com/AlexBlackNn/metrics/internal/config/configagent"
	"github.com/AlexBlackNn/metrics/internal/config/configserver"
	"github.com/AlexBlackNn/metrics/internal/domain/models"
	"github.com/AlexBlackNn/metrics/internal/services/agentmetricsservice"
	"golang.org/x/time/rate"
)

type Sender struct {
	log       *slog.Logger
	cfg       *configagent.Config
	encryptor *encryption.Encryptor
	*agentmetricsservice.MonitorService
}

func New(
	log *slog.Logger,
	cfg *configagent.Config,
	encryptor *encryption.Encryptor,
) *Sender {

	return &Sender{
		log,
		cfg,
		encryptor,
		agentmetricsservice.New(log, cfg),
	}
}

// getLocalIP returns the non loopback local IP of the host
func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				fmt.Println("ipnet", ipnet.IP.String())
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

func (s *Sender) Send(ctx context.Context) {

	log := s.log.With(
		slog.String("info", "SERVICE LAYER: metricsHttpService.Transmit"),
	)
	reportInterval := time.Duration(s.cfg.ReportInterval) * time.Second
	rateLimiter := rate.NewLimiter(rate.Limit(s.cfg.AgentRateLimit), s.cfg.AgentBurstTokens)

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(reportInterval):
			for _, savedMetric := range s.GetMetrics() {
				err := rateLimiter.Wait(ctx)
				if err != nil {
					log.Error(err.Error())
					return
				}
				go func(savedMetric models.MetricInteraction) {
					restyClient := resty.New()
					restyClient.
						SetRetryCount(s.cfg.AgentRetryCount).
						SetRetryWaitTime(s.cfg.AgentRetryWaitTime).
						SetRetryMaxWaitTime(s.cfg.AgentRetryMaxWaitTime)

					var body string
					if savedMetric.GetType() == configserver.MetricTypeCounter {
						body = fmt.Sprintf(`{"id":"%s", "type":"%s", "delta": %d}`,
							savedMetric.GetName(),
							savedMetric.GetType(),
							savedMetric.GetValue(),
						)
					} else {
						body = fmt.Sprintf(`{"id":"%s", "type":"%s", "value": %v}`,
							savedMetric.GetName(),
							savedMetric.GetType(),
							savedMetric.GetValue(),
						)
					}

					body, err = s.encryptor.EncryptMessage(body)
					if err != nil {
						log.Error("error creating encrypted message", "error", err.Error())
						return
					}
					hashCalculator := hash.New(s.cfg)
					hashResult := hashCalculator.MetricHash(body)

					url := fmt.Sprintf("http://%s/update/", s.cfg.ServerAddr)
					log.Info("sending data", "url", url)
					var resp *resty.Response
					if s.cfg.CryptoKeyPath != "" {
						resp, err = restyClient.R().
							SetHeader("Content-Type", "application/json").
							SetHeader("HashSHA256", hashResult).
							SetHeader("X-Encrypted", "true").
							SetHeader("X-Real-IP", getLocalIP()).
							SetHeader("X-Encryption-Method", "RSA").
							SetBody(body).
							Post(url)
					} else {
						resp, err = restyClient.R().
							SetHeader("Content-Type", "application/json").
							SetHeader("HashSHA256", hashResult).
							SetHeader("X-Real-IP", getLocalIP()).
							SetBody(body).
							Post(url)
					}
					if err != nil {
						log.Error("error creating http request")
						return
					}
					log.Info("http request finished successfully",
						"url", url,
						"statusCode", resp.StatusCode(),
						"body", string(resp.Body()),
					)
				}(savedMetric)
			}
		}
	}
}
