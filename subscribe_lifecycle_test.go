package signals_test

import (
	"context"
	"runtime"
	"sync"
	"testing"

	"github.com/coregx/signals"
)

// TestSubscribeForever_10kHeldOpen_NoPerSubscriptionGoroutine is the
// goroutine-leak regression test.
//
// A "subscribe then immediately unsubscribe" loop is not a meaningful
// regression test here: the pre-fix implementation spawns a goroutine per
// Subscribe call, but that goroutine correctly exits as soon as
// Unsubscribe is called manually — sequential subscribe+unsubscribe
// passes on both the old and new implementation. The actual defect is the
// goroutine parked for the entire time a subscription is held open
// (SubscribeForever passes context.Background(), which never cancels, so
// the old implementation's goroutine sits blocked on select for as long
// as the subscription lives — potentially the process lifetime).
//
// So this holds 10k subscriptions open simultaneously and asserts
// NumGoroutine does NOT grow anywhere near 10k while they're all live —
// the context.AfterFunc-based implementation registers no goroutine at
// all for context.Background().
func TestSubscribeForever_10kHeldOpen_NoPerSubscriptionGoroutine(t *testing.T) {
	const n = 10_000

	sig := signals.New(0)

	runtime.GC()
	runtime.Gosched()
	baseline := runtime.NumGoroutine()

	unsubs := make([]signals.Unsubscribe, n)
	for i := range n {
		unsubs[i] = sig.SubscribeForever(func(int) {})
	}

	runtime.Gosched()
	held := runtime.NumGoroutine() - baseline
	if held > n/10 {
		t.Errorf("NumGoroutine grew by %d while %d subscriptions were held open, want << %d (goroutine parked per Subscribe)",
			held, n, n)
	}

	for _, unsub := range unsubs {
		unsub()
	}

	runtime.GC()
	runtime.Gosched()
	delta := runtime.NumGoroutine() - baseline
	if delta > 10 {
		t.Errorf("NumGoroutine delta = %d after unsubscribing all %d, want <= 10", delta, n)
	}
}

// TestSubscribe_ConcurrentCancelAndUnsubscribe_NoRace pins the
// close-close race: the pre-fix implementation raced two different
// close(done) call sites (context cancellation vs. manual Unsubscribe)
// guarded only by a non-atomic select-with-default, which could panic
// with "close of closed channel" when both fired at once. The
// context.AfterFunc + sync.Once rewrite must make concurrent
// cancel+unsubscribe race-free. Run with -race to also catch any data
// race, not just panics.
func TestSubscribe_ConcurrentCancelAndUnsubscribe_NoRace(t *testing.T) {
	const cycles = 10_000

	sig := signals.New(0)

	var wg sync.WaitGroup
	for range cycles {
		ctx, cancel := context.WithCancel(context.Background())
		unsub := sig.Subscribe(ctx, func(int) {})

		wg.Go(cancel)
		wg.Go(unsub)
	}
	wg.Wait()
}
