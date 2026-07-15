package signals

import (
	"context"
	"testing"
)

// Allocation regression tests.
// These use testing.AllocsPerRun which is deterministic —
// unlike benchmark ns/op, alloc counts don't fluctuate,
// so any increase immediately fails the build.

func TestAllocations_Signal_SubscribeForever(t *testing.T) {
	sig := New(0)
	allocs := testing.AllocsPerRun(100, func() {
		unsub := sig.SubscribeForever(func(int) {})
		unsub()
	})
	// 2 allocs: atomic.Bool escape + remove closure
	if allocs > 2 {
		t.Errorf("SubscribeForever: %.0f allocs, want <= 2", allocs)
	}
}

func TestAllocations_Signal_SubscribeWithContext(t *testing.T) {
	sig := New(0)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	allocs := testing.AllocsPerRun(100, func() {
		unsub := sig.Subscribe(ctx, func(int) {})
		unsub()
	})
	// 5 allocs: atomic.Bool + remove closure + wrapper closure + AfterFunc internals
	if allocs > 5 {
		t.Errorf("Subscribe(ctx): %.0f allocs, want <= 5", allocs)
	}
}

func TestAllocations_Computed_SubscribeForever(t *testing.T) {
	base := New(0)
	comp := Computed(func() int { return base.Get() * 2 }, base.AsReadonly())

	allocs := testing.AllocsPerRun(100, func() {
		unsub := comp.SubscribeForever(func(int) {})
		unsub()
	})
	if allocs > 2 {
		t.Errorf("Computed.SubscribeForever: %.0f allocs, want <= 2", allocs)
	}
}

func TestAllocations_Signal_Get(t *testing.T) {
	sig := New(42)
	allocs := testing.AllocsPerRun(100, func() {
		_ = sig.Get()
	})
	if allocs > 0 {
		t.Errorf("Get: %.0f allocs, want 0", allocs)
	}
}

func TestAllocations_Signal_Set(t *testing.T) {
	sig := New(0)
	allocs := testing.AllocsPerRun(100, func() {
		sig.Set(1)
	})
	// Set with no subscribers = 0 allocs (no callback slice created)
	if allocs > 0 {
		t.Errorf("Set (no subscribers): %.0f allocs, want 0", allocs)
	}
}

func TestAllocations_Signal_SetWithSubscribers(t *testing.T) {
	sig := New(0)
	sig.SubscribeForever(func(int) {})

	allocs := testing.AllocsPerRun(100, func() {
		sig.Set(1)
	})
	// Set with subscribers: 1 alloc for callback slice copy
	if allocs > 1 {
		t.Errorf("Set (with subscriber): %.0f allocs, want <= 1", allocs)
	}
}

func TestAllocations_Computed_Get_Cached(t *testing.T) {
	base := New(42)
	comp := Computed(func() int { return base.Get() * 2 }, base.AsReadonly())
	_ = comp.Get() // prime cache

	allocs := testing.AllocsPerRun(100, func() {
		_ = comp.Get()
	})
	if allocs > 0 {
		t.Errorf("Computed.Get (cached): %.0f allocs, want 0", allocs)
	}
}

func TestAllocations_AsReadonly(t *testing.T) {
	sig := New(42)
	_ = sig.AsReadonly() // prime cache

	allocs := testing.AllocsPerRun(100, func() {
		_ = sig.AsReadonly()
	})
	// Cached: 0 allocs after first call
	if allocs > 0 {
		t.Errorf("AsReadonly (cached): %.0f allocs, want 0", allocs)
	}
}

func TestAllocations_Effect_Create(t *testing.T) {
	base := New(0)
	readonly := base.AsReadonly()

	allocs := testing.AllocsPerRun(20, func() {
		eff := Effect(func() {
			_ = base.Get()
		}, readonly)
		eff.Stop()
	})
	// 7 allocs: effect struct, pre-alloc slice, method value,
	// type-erased callback, atomic.Bool, remove closure, make(slice)
	if allocs > 7 {
		t.Errorf("Effect (1 dep): %.0f allocs, want <= 7", allocs)
	}
}

func TestAllocations_EffectWithCleanup_Create(t *testing.T) {
	base := New(0)
	readonly := base.AsReadonly()

	allocs := testing.AllocsPerRun(20, func() {
		eff := EffectWithCleanup(func() func() {
			_ = base.Get()
			return func() {}
		}, readonly)
		eff.Stop()
	})
	// 7 allocs: same as Effect but fn instead of fnSimple
	if allocs > 7 {
		t.Errorf("EffectWithCleanup (1 dep): %.0f allocs, want <= 7", allocs)
	}
}
