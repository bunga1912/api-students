package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hibiken/asynq"

	"api-students/app/model"
	"api-students/app/repository"
	"api-students/helper"
)

// BuildQuotaMessage menyusun email "kuota berubah" untuk satu penerima.
// Fungsi murni, seperti BuildMessage.
func BuildQuotaMessage(e model.QuotaChangedEvent, user model.User) model.Message {
	course := e.CourseCode + " " + e.CourseName
	return model.Message{
		To:      user.Email,
		Subject: fmt.Sprintf("KRS: kuota %s berubah menjadi %d", course, e.Quota),
		Body: fmt.Sprintf("Halo %s, kuota kelas %s kini menjadi %d.",
			user.Username, course, e.Quota),
	}
}

// deliver mengirim satu email dengan pola yang sama seperti HandleSend:
// claim dan pengiriman berada di dalam SATU transaction, dan claim
// dilakukan lewat repository hasil WithTx(tx).
//
// Kuncinya adalah (outbox_id, user_id): bila attempt berikutnya datang
// setelah email ke penerima ini sudah terkirim, claim ditolak dan email
// dilewati. Itulah yang mencegah email ke-1 sampai ke-6 terkirim ulang
// ketika email ke-7 gagal dan task di-retry.
func (w *NotificationWorker) deliver(
	ctx context.Context,
	outboxID int64,
	user model.User,
	msg model.Message,
) (alreadySent bool, err error) {
	err = w.tx.WithTxOnce(ctx, func(tx repository.DBTX) error {
		claimed, err := w.notifications.WithTx(tx).Claim(ctx, model.Notification{
			OutboxID: outboxID, UserID: user.ID, Channel: "email",
			Subject: msg.Subject, Body: msg.Body,
		})
		if err != nil {
			return err
		}
		if !claimed {
			alreadySent = true
			return nil
		}
		return w.mailer.Send(ctx, msg)
	})
	return alreadySent, err
}

// handleQuotaChanged mengirim email ke seluruh peserta kelas.
//
// Satu kegagalan sementara tidak menghentikan penerima lain; task
// dikembalikan sebagai error di akhir supaya Asynq melakukan retry, dan
// retry itu hanya akan mengirim ulang yang belum terkirim.
func (w *NotificationWorker) handleQuotaChanged(ctx context.Context, job model.NotificationJob) error {
	var ev model.QuotaChangedEvent
	if err := json.Unmarshal(job.Payload, &ev); err != nil {
		return fmt.Errorf("%w: payload event %d rusak: %v", asynq.SkipRetry, job.OutboxID, err)
	}

	recipients, err := w.users.FindByCourse(ctx, ev.CourseID)
	if err != nil {
		return err // database bermasalah: layak di-retry
	}

	retry, _ := asynq.GetRetryCount(ctx)
	failed := 0
	var lastErr error

	for _, u := range recipients {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		alreadySent, err := w.deliver(ctx, job.OutboxID, u, BuildQuotaMessage(ev, u))
		switch {
		case errors.Is(err, helper.ErrPermanentDelivery):
			// Alamat penerima ini tidak ada. Melakukan retry tidak membantu,
			// dan tidak boleh menahan penerima lain.
			w.logger.Warn("quota_recipient_skipped",
				"outbox_id", job.OutboxID, "user_id", u.ID, "error", err.Error())
			continue
		case err != nil:
			failed++
			lastErr = err
			continue
		}

		w.logger.Info("notification_processed",
			"outbox_id", job.OutboxID,
			"event", job.EventType,
			"user_id", u.ID,
			"retry", retry,
			"already_sent", alreadySent)
	}

	if failed > 0 {
		return fmt.Errorf("%d dari %d email gagal terkirim: %w", failed, len(recipients), lastErr)
	}
	return nil
}