package channel

import (
	b64 "encoding/base64"
	"testing"

	notifyingv1alpha1 "github.com/containers-ai/alameda/notifier/api/v1alpha1"
)

func Test_SendEmailBySMTP(t *testing.T) {
	type args struct {
		notificationChannel *notifyingv1alpha1.AlamedaNotificationChannel
		To                  []string
		Cc                  []string
		Subject             string
		Msg                 string
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
				To:          []string{""},
				Cc:          []string{""},
				Subject:     "test TLS subject",
				Msg:         "testing TLS connection",
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
				To:          []string{""},
				Cc:          []string{""},
				Subject:     "test STARTTLS subject",
				Msg:         "testing STARTTLS connection",
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
			subject := tt.args.Subject
			from := tt.args.notificationChannel.Spec.Email.From
			recipients := tt.args.To
			ccs := tt.args.Cc
			msg := tt.args.Msg
			attachments := tt.args.Attachments
			emailClient := EmailClient{
				client: got,
			}
			err = emailClient.SendEmailBySMTP(subject, from, recipients, msg, ccs, attachments)
			if (err != nil) != tt.wantErr {
				t.Errorf("SendEmailBySMTP() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
