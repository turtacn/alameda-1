package channel

import (
	"crypto/tls"
	"encoding/base64"
	b64 "encoding/base64"
	"fmt"
	"io/ioutil"
	"net/smtp"
	"strings"

	notifyingv1alpha1 "github.com/containers-ai/alameda/notifier/api/v1alpha1"
	notifier_utils "github.com/containers-ai/alameda/notifier/utils"
	"github.com/containers-ai/alameda/pkg/utils/log"
	datahub_events "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/events"
	"github.com/pkg/errors"
	"gopkg.in/mail.v2"
)

var scope = log.RegisterScope("email", "email", 0)

type EmailClient struct {
	notificationChannel *notifyingv1alpha1.AlamedaNotificationChannel
	emailChannel        *notifyingv1alpha1.AlamedaEmailChannel
	client              interface{}
	auth                smtp.Auth
	mailAddr            string
	clusterInfo         *notifier_utils.ClusterInfo
}

func NewEmailClient(notificationChannel *notifyingv1alpha1.AlamedaNotificationChannel,
	emailChannel *notifyingv1alpha1.AlamedaEmailChannel, clusterInfo *notifier_utils.ClusterInfo) (*EmailClient, error) {
	host := notificationChannel.Spec.Email.Server
	port := notificationChannel.Spec.Email.Port

	client, err := getSMTPClient(notificationChannel)

	if err != nil {
		return nil, err
	}

	return &EmailClient{
		notificationChannel: notificationChannel,
		emailChannel:        emailChannel,
		mailAddr:            fmt.Sprintf("%s:%v", host, port),
		client:              client,
		clusterInfo:         clusterInfo,
	}, nil
}

func (emailClient *EmailClient) SendEvent(evt *datahub_events.Event) error {
	msg := evt.GetMessage()
	subject := notifier_utils.EventEmailSubject(evt)
	from := emailClient.notificationChannel.Spec.Email.From
	recipients := emailClient.emailChannel.To
	ccs := emailClient.emailChannel.Cc
	// key/value -> filename/filepath
	attachments := map[string]string{}
	scope.Infof("Start sending email (subject: %s, from: %s, to: %s, cc:%s, body: %s)",
		subject, from, strings.Join(recipients, ";"), strings.Join(ccs, ";"), msg)
	err := emailClient.SendEmailBySMTP(subject, from, recipients, notifier_utils.EventHTMLMsg(evt, emailClient.clusterInfo),
		notifier_utils.RemoveEmptyStr(ccs), attachments)
	if err != nil {
		return err
	}
	return nil
}

func (emailClient *EmailClient) SendEmailBySMTP(subject string, from string,
	recipients []string, msgHTML string, ccs []string, attachments map[string]string) error {

	if client, ok := emailClient.client.(*smtp.Client); ok {
		if err := client.Mail(from); err != nil {
			return errors.Wrap(err,
				"issue MAIL command for the provided email address failed")
		}
		for _, recipient := range recipients {
			if err := client.Rcpt(recipient); err != nil {
				return errors.Wrap(err,
					"issue RCPT command for provided email addresses failed")
			}
		}
		for _, cc := range ccs {
			if err := client.Rcpt(cc); err != nil {
				return errors.Wrap(err,
					"issue RCPT command for provided email addresses (CC) failed")
			}
		}
		wc, err := client.Data()
		if err != nil {
			return errors.Wrap(err, "issue DATA command failed")
		}
		sentBody := getBodyString(subject, from, recipients, msgHTML, ccs, attachments)
		_, err = fmt.Fprintf(wc, sentBody)
		if err != nil {
			return errors.Wrap(err, "email body format failed")
		}

		err = wc.Close()
		if err != nil {
			return errors.Wrap(err, "close email writer failed")
		}

		// Send the QUIT command and close the connection.
		err = client.Quit()
		if err != nil {
			return errors.Wrap(err, "send email QUIT command failed")
		}
	} else if client, ok := emailClient.client.(*mail.Dialer); ok {
		mailMsg := getMailMessage(subject, from, recipients, msgHTML, ccs, attachments)
		return client.DialAndSend(mailMsg)
	}
	return nil
}

func getMailMessage(subject string, from string, recipients []string,
	msgHTML string, ccs []string, attachments map[string]string) *mail.Message {
	mailMsg := mail.NewMessage()
	mailMsg.SetHeaders(
		map[string][]string{
			"From":    []string{from},
			"To":      recipients,
			"Cc":      ccs,
			"Subject": []string{subject},
		})
	mailMsg.SetBody("text/html", msgHTML)
	for _, filePath := range attachments {
		mailMsg.Attach(filePath)
	}
	return mailMsg
}

func getBodyString(subject string, from string, recipients []string,
	msgHTML string, ccs []string, attachments map[string]string) string {

	sentBody := fmt.Sprintf("To: %s\r\n", strings.Join(recipients, ";"))
	if len(ccs) > 0 {
		sentBody = fmt.Sprintf("%sCc: %s\r\n", sentBody, strings.Join(ccs, ";"))
	}
	sentBody = fmt.Sprintf("%sSubject: %s\r\n", sentBody, subject)
	mimeVer := "1.0"
	sentBody = fmt.Sprintf("%sMIME-Version: %s\r\n", sentBody, mimeVer)
	delimeter := "----=_NextPart_ProhetStor_888"
	sentBody = fmt.Sprintf("%sContent-Type: multipart/mixed; boundary=\"%s\"\r\n",
		sentBody, delimeter)
	sentBody = fmt.Sprintf("%s\r\n--%s\r\n",
		sentBody, delimeter)
	sentBody = fmt.Sprintf("%sContent-Type: text/html; charset=\"utf-8\"\r\nContent-Transfer-Encoding: 7bit\r\n",
		sentBody)
	sentBody = fmt.Sprintf("%s\r\n%s\r\n",
		sentBody, msgHTML)

	// attachments

	for fileName, filePath := range attachments {
		sentBody = fmt.Sprintf("%s\r\n--%s\r\n", sentBody, delimeter)
		sentBody = fmt.Sprintf("%sContent-Type: text/plain; charset=\"utf-8\"\r\nContent-Transfer-Encoding: base64\r\n",
			sentBody)
		sentBody = fmt.Sprintf("%sContent-Disposition: attachment;filename=\"%s\"\r\n",
			sentBody, fileName)
		rawFile, err := ioutil.ReadFile(filePath)
		if err != nil {
			scope.Errorf(err.Error())
		}
		attachMsg := base64.StdEncoding.EncodeToString(rawFile)
		sentBody = fmt.Sprintf("%s\r\n%s", sentBody, attachMsg)
	}
	return sentBody
}

func getSMTPClient(notificationChannel *notifyingv1alpha1.AlamedaNotificationChannel) (interface{}, error) {
	host := notificationChannel.Spec.Email.Server
	port := notificationChannel.Spec.Email.Port
	addr := fmt.Sprintf("%s:%v", host, port)
	usernameBin, err := b64.StdEncoding.DecodeString(notificationChannel.Spec.Email.Username)
	if err != nil {
		return nil, errors.Wrap(err, "decode username failed")
	}
	passwordBin, err := b64.StdEncoding.DecodeString(notificationChannel.Spec.Email.Password)
	if err != nil {
		return nil, errors.Wrap(err, "decode password failed")
	}
	username := string(usernameBin)
	password := string(passwordBin)
	encryption := notificationChannel.Spec.Email.Encryption
	auth := smtp.PlainAuth("", username, password, host)
	if strings.ToLower(encryption) == "starttls" {
		d := mail.NewDialer(host, int(port), username, password)
		d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
		d.StartTLSPolicy = mail.MandatoryStartTLS
		return d, nil
		// office365 does not receive mail from built-in smtp library
		/*
			tlsconfig := &tls.Config{
				InsecureSkipVerify: true,
				ServerName:         addr,
			}
			client, err := smtp.Dial(addr)
			if err != nil {
				return nil, errors.Wrap(err,
					"create smtp client using STARTTLS encryption failed")
			}
			err = client.StartTLS(tlsconfig)
			if err != nil {
				return client, errors.Wrap(err,
					"create smtp client using STARTTLS encryption with auth failed")
			}
			err = client.Auth(auth)
			if err != nil {
				return client, errors.Wrap(err,
					"create smtp client using STARTTLS encryption with auth failed")
			}
			return client, nil
		*/
	} else {
		/* default is tls/ssl encryption*/
		tlsconfig := &tls.Config{
			InsecureSkipVerify: true,
			ServerName:         addr,
		}
		conn, err := tls.Dial("tcp", addr, tlsconfig)
		if err != nil {
			return nil, errors.Wrap(err,
				"tls dial failed")
		}
		client, err := smtp.NewClient(conn, host)
		if err != nil {
			return nil, errors.Wrap(err,
				"create smtp client using SSL/TLS encryption failed")
		}
		err = client.Auth(auth)
		if err != nil {
			return client, errors.Wrap(err,
				"create smtp client using SSL/TLS encryption with auth failed")
		}
		return client, nil
	}
}
