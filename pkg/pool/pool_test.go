// pool_test.go
package pool

import (
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// ────────────────────────────────────────────────
// Тестовая структура, реализующая Resetter

type TestItem struct {
	ID      int64
	Counter int32    // будет увеличиваться при использовании
	Dirty   bool     // флаг, который должен сбрасываться в Reset
	Extra   [16]byte // немного памяти, чтобы было интереснее
}

func (t *TestItem) Reset() {
	atomic.StoreInt32(&t.Counter, 0)
	t.Dirty = false
}

func NewTestItem(id int64) *TestItem {
	return &TestItem{
		ID:      id,
		Counter: 0,
		Dirty:   true, // изначально "грязный"
	}
}

// simulateWork имитирует использование объекта
func (t *TestItem) simulateWork() {
	atomic.AddInt32(&t.Counter, 1)
	t.Dirty = true
	time.Sleep(time.Microsecond * time.Duration(10+rand.Intn(40)))
}

// ────────────────────────────────────────────────
// Тесты

func TestPool_BasicGetPutReset(t *testing.T) {
	p := New(func() *TestItem {
		return NewTestItem(42)
	})

	item1 := p.Get()
	if item1 == nil {
		t.Fatal("первый Get не должен возвращать nil")
	}
	if item1.Counter != 0 {
		t.Errorf("новый объект должен иметь Counter=0, получено %d", item1.Counter)
	}

	item1.simulateWork()

	if atomic.LoadInt32(&item1.Counter) == 0 {
		t.Error("после работы Counter должен быть > 0")
	}
	if !item1.Dirty {
		t.Error("после работы Dirty должен быть true")
	}

	p.Put(item1)

	item2 := p.Get()
	if item2 == nil {
		t.Fatal("после Put должен вернуться существующий объект")
	}

	if item2 != item1 {
		t.Error("должен вернуться тот же самый указатель")
	}
	if atomic.LoadInt32(&item2.Counter) != 0 {
		t.Error("после Reset Counter должен быть 0")
	}
	if item2.Dirty {
		t.Error("после Reset Dirty должен быть false")
	}
}

func TestPool_MultipleObjects(t *testing.T) {
	p := New(func() *TestItem {
		return NewTestItem(rand.Int63())
	})

	const count = 50
	items := make([]*TestItem, count)

	for i := 0; i < count; i++ {
		items[i] = p.Get()
		items[i].simulateWork()
	}

	// возвращаем все обратно
	for _, it := range items {
		p.Put(it)
	}

	// берём заново и проверяем сброс
	for i := 0; i < count; i++ {
		it := p.Get()
		if atomic.LoadInt32(&it.Counter) != 0 {
			t.Errorf("элемент #%d: Counter не сброшен: %d", i, it.Counter)
		}
		if it.Dirty {
			t.Errorf("элемент #%d: Dirty не сброшен", i)
		}
	}
}

func TestPool_Concurrency(t *testing.T) {
	p := New(func() *TestItem {
		return NewTestItem(rand.Int63())
	})

	const workers = 12
	const iterations = 400

	var wg sync.WaitGroup
	wg.Add(workers)

	var gotFromPool atomic.Int32
	var createdNew atomic.Int32

	for w := 0; w < workers; w++ {
		go func(workerID int) {
			defer wg.Done()
			rng := rand.New(rand.NewSource(time.Now().UnixNano() + int64(workerID)))

			for i := 0; i < iterations; i++ {
				item := p.Get()

				wasNew := atomic.LoadInt32(&item.Counter) == 0 && !item.Dirty
				if !wasNew {
					gotFromPool.Add(1)
				} else {
					createdNew.Add(1)
				}

				item.simulateWork()

				// случайная задержка перед возвратом
				time.Sleep(time.Microsecond * time.Duration(50+rng.Intn(300)))

				p.Put(item)
			}
		}(w)
	}

	wg.Wait()

	t.Logf("Получено из пула: %d", gotFromPool.Load())
	t.Logf("Создано новых:   %d", createdNew.Load())
	t.Logf("Всего операций:  %d", workers*iterations)

	if gotFromPool.Load() == 0 && createdNew.Load() > 0 {
		t.Error("в конкурентном режиме хотя бы часть объектов должна была быть взята из пула повторно")
	}
}

func TestPool_ZeroValueBehavior(t *testing.T) {
	p := New(func() *TestItem {
		return NewTestItem(100)
	})

	// специально не кладём ничего в пул заранее
	item := p.Get()

	if item == nil {
		t.Fatal("Get не должен возвращать nil даже при пустом пуле")
	}

	// имитируем использование и кладём обратно
	item.simulateWork()
	p.Put(item)

	itemAgain := p.Get()
	if atomic.LoadInt32(&itemAgain.Counter) != 0 {
		t.Error("после Put должен сработать Reset")
	}
	if itemAgain.Dirty {
		t.Error("после Reset Dirty должен быть false")
	}
}
