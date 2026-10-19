package batch_mail

import (
	"billionmail-core/internal/service/public"
	"context"
	"sync"
	"time"
)

var nativeSendSlots struct {
	sync.Mutex
	next time.Time
}

// All tasks share the imported YXHK sending interval in native mode.
func waitNativeSendSlot(ctx context.Context) error {
	interval := 20
	_ = public.OptionsMgrInstance.GetOption(ctx, "native_send_interval", &interval)
	if interval < 1 {
		interval = 20
	}
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		nativeSendSlots.Lock()
		wait := time.Until(nativeSendSlots.next)
		if wait <= 0 {
			nativeSendSlots.next = time.Now().Add(time.Duration(interval) * time.Second)
			nativeSendSlots.Unlock()
			return nil
		}
		nativeSendSlots.Unlock()
		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
}
