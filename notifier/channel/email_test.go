package channel

import (
	b64 "encoding/base64"
	"testing"
	"time"

	notifyingv1alpha1 "github.com/containers-ai/alameda/notifier/api/v1alpha1"
	"github.com/containers-ai/alameda/notifier/utils"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/spf13/viper"
)

func Test_SendEmailBySMTP(t *testing.T) {
	configFile := "/etc/alameda/notifier/notifier.toml"
	viper.SetConfigFile(configFile)

	if err := viper.ReadInConfig(); err != nil {
		t.Errorf("getSMTPClient() error = %v", err)
	}
	type args struct {
		notificationChannel *notifyingv1alpha1.AlamedaNotificationChannel
		To                  []string
		Cc                  []string
		Event               *datahub_v1alpha1.Event
		Attachments         map[string]string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "tls",
			args: args{
				notificationChannel: &notifyingv1alpha1.AlamedaNotificationChannel{
					Spec: notifyingv1alpha1.AlamedaNotificationChannelSpec{
						Type: "email",
						Email: notifyingv1alpha1.AlamedaEmail{
							Server:     "172.31.1.1",
							Port:       465,
							From:       "",
							Username:   b64.StdEncoding.EncodeToString([]byte("")),
							Password:   b64.StdEncoding.EncodeToString([]byte("")),
							Encryption: "tls",
						},
					},
				},
				To: []string{""},
				Cc: []string{""},
				Event: &datahub_v1alpha1.Event{
					Time: &timestamp.Timestamp{
						Seconds: time.Now().Unix(),
					},
					Source: &datahub_v1alpha1.EventSource{
						Host:      "email warning host",
						Component: "email warning component",
					},
					Type:    datahub_v1alpha1.EventType_EVENT_TYPE_EMAIL_NOTIFICATION,
					Version: datahub_v1alpha1.EventVersion_EVENT_VERSION_V1,
					Level:   datahub_v1alpha1.EventLevel_EVENT_LEVEL_WARNING,
					Subject: &datahub_v1alpha1.K8SObjectReference{
						Kind:      "Pod",
						Namespace: "federatorai",
						Name:      "alameda-notifier-7d6b779c47-f6t7q",
					},
					Message: "send email warning message",
					Data:    "send email warning data",
				},
				Attachments: map[string]string{
					// filename: filepath
				},
			},
			wantErr: false,
		},
		{
			name: "starttls",
			args: args{
				notificationChannel: &notifyingv1alpha1.AlamedaNotificationChannel{
					Spec: notifyingv1alpha1.AlamedaNotificationChannelSpec{
						Type: "email",
						Email: notifyingv1alpha1.AlamedaEmail{
							Server:     "smtp.office365.com",
							Port:       587,
							From:       "",
							Username:   b64.StdEncoding.EncodeToString([]byte("")),
							Password:   b64.StdEncoding.EncodeToString([]byte("")),
							Encryption: "starttls",
						},
					},
				},
				To: []string{""},
				Cc: []string{""},
				Event: &datahub_v1alpha1.Event{
					Time: &timestamp.Timestamp{
						Seconds: time.Now().Unix(),
					},
					Source: &datahub_v1alpha1.EventSource{
						Host:      "email warning host",
						Component: "email warning component",
					},
					Type:    datahub_v1alpha1.EventType_EVENT_TYPE_EMAIL_NOTIFICATION,
					Version: datahub_v1alpha1.EventVersion_EVENT_VERSION_V1,
					Level:   datahub_v1alpha1.EventLevel_EVENT_LEVEL_WARNING,
					Subject: &datahub_v1alpha1.K8SObjectReference{
						Kind:      "Pod",
						Namespace: "federatorai",
						Name:      "alameda-notifier-7d6b779c47-f6t7q",
					},
					Message: "send email warning message",
					Data:    "send email warning data",
				},
				Attachments: map[string]string{
					// filename: filepath
				},
			},
			wantErr: false,
		},
	}
	for _, ttt := range tests {
		tt := ttt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := getSMTPClient(tt.args.notificationChannel)
			if (err != nil) != tt.wantErr {
				t.Errorf("getSMTPClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			subject := utils.EventEmailSubject(tt.args.Event)
			from := tt.args.notificationChannel.Spec.Email.From
			recipients := tt.args.To
			ccs := tt.args.Cc
			msgHTML := utils.EventHTMLMsg(tt.args.Event)
			attachments := tt.args.Attachments
			emailClient := EmailClient{
				client: got,
			}
			err = emailClient.SendEmailBySMTP(subject, from, recipients, msgHTML, utils.RemoveEmptyStr(ccs), attachments)
			if (err != nil) != tt.wantErr {
				t.Errorf("SendEmailBySMTP() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
