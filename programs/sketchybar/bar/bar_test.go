package bar

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

type fakeRunner struct {
	mu      sync.Mutex
	outputs map[string]string
	calls   [][]string
}

func newFakeRunner(outputs map[string]string) *fakeRunner {
	return &fakeRunner{outputs: outputs}
}

func (f *fakeRunner) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, append([]string{name}, args...))
	key := strings.Join(append([]string{name}, args...), " ")
	for prefix, out := range f.outputs {
		if strings.HasPrefix(key, prefix) {
			return []byte(out), nil
		}
	}
	if name == "sketchybar" {
		return nil, nil
	}
	return nil, fmt.Errorf("unexpected command: %s", key)
}

func (f *fakeRunner) sketchybarCalls() [][]string {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out [][]string
	for _, c := range f.calls {
		if c[0] == "sketchybar" && !slices.Equal(c, []string{"sketchybar", "--query", "bar"}) {
			out = append(out, c[1:])
		}
	}
	return out
}

func containsSeq(haystack []string, needle ...string) bool {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if slices.Equal(haystack[i:i+len(needle)], needle) {
			return true
		}
	}
	return false
}

func TestWorkspaceShowsFocused(t *testing.T) {
	r := newFakeRunner(map[string]string{"aerospace list-workspaces --focused": "3\n"})
	w := NewWorkspace(r, Colors{})
	b := &Batch{}
	if err := w.Setup(context.Background(), b); err != nil {
		t.Fatal(err)
	}
	if !containsSeq(b.Args(), "--add", "event", WorkspaceChangeEvent) || !slices.Contains(b.Args(), "icon="+iconFlower) {
		t.Fatalf("unexpected setup: %v", b.Args())
	}

	b = &Batch{}
	w.Handle(context.Background(), Event{Env: map[string]string{"FOCUSED_WORKSPACE": "T"}}, b)
	if !slices.Equal(b.Args(), []string{"--set", "space", "label=T"}) {
		t.Fatalf("FOCUSED_WORKSPACE should win: %v", b.Args())
	}

	b = &Batch{}
	w.Handle(context.Background(), Event{}, b)
	if !slices.Equal(b.Args(), []string{"--set", "space", "label=3"}) {
		t.Fatalf("should fall back to aerospace: %v", b.Args())
	}
}

func TestAppIcon(t *testing.T) {
	cases := map[string]string{
		"Firefox":          ":firefox:",
		"kitty":            ":kitty:",
		"Proton Mail":      ":proton_mail:",
		"MAIL":             ":mail:",
		"Some Unknown App": appIconDefault,
	}
	for app, want := range cases {
		if got := appIcon(app); got != want {
			t.Errorf("appIcon(%q) = %q, want %q", app, got, want)
		}
	}
}

func TestParseStatusLabel(t *testing.T) {
	cases := map[string]string{
		`"StatusLabel"={ "label"="12" }`: "12",
		`"StatusLabel"={ "label"="" }`:   "",
		``:                               "",
	}
	for in, want := range cases {
		if got := parseStatusLabel(in); got != want {
			t.Errorf("parseStatusLabel(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestBadges(t *testing.T) {
	r := newFakeRunner(map[string]string{
		"lsappinfo info -only StatusLabel Proton Mail": `"StatusLabel"={ "label"="" }`,
		"lsappinfo info -only StatusLabel Discord":     `"StatusLabel"={ "label"="2" }`,
	})
	mood := NewMood()
	bg := NewBadges(r, Colors{}, []string{"Proton Mail", "Discord"}, mood)

	b := &Batch{}
	if err := bg.Setup(context.Background(), b); err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(b.Args(), "icon=:discord:") || !containsSeq(b.Args(), "--add", "bracket", badgesBracket) {
		t.Fatalf("unexpected setup: %v", b.Args())
	}

	b = &Batch{}
	bg.Poll(context.Background())(b)
	args := b.Args()
	for _, want := range [][]string{
		{"--set", "badge.0", "drawing=off"},
		{"--set", "badge.1", "drawing=on"},
		{"--set", "badge.1.count", "drawing=on", "label=2"},
		{"--set", badgesBracket, "drawing=on"},
	} {
		if !containsSeq(args, want...) {
			t.Errorf("missing %v in %v", want, args)
		}
	}
	if mood.Face() != faceSparkly && time.Now().Hour() >= 6 {
		t.Errorf("new unread should make the face sparkly, got %q", mood.Face())
	}
	if bg.Poll(context.Background()) != nil {
		t.Error("unchanged badges should not produce an update")
	}
}

func TestRenderBadgesDotAndEmpty(t *testing.T) {
	b := &Batch{}
	renderBadges(b, []string{"•", ""})
	if !containsSeq(b.Args(), "--set", "badge.0.count", "drawing=on", "label=", "background.height=8") {
		t.Fatalf("dot badge should render as a small dot: %v", b.Args())
	}

	b = &Batch{}
	renderBadges(b, []string{"", "0"})
	if !containsSeq(b.Args(), "--set", badgesBracket, "drawing=off") {
		t.Fatalf("island should hide when nothing is unread: %v", b.Args())
	}
}

func TestUnreadTotal(t *testing.T) {
	if got := unreadTotal([]string{"3", "•", "", "x"}); got != 4 {
		t.Fatalf("unreadTotal = %d, want 4", got)
	}
}

func TestMoodPriority(t *testing.T) {
	at := func(hour int) func() time.Time {
		return func() time.Time { return time.Date(2026, 10, 13, hour, 0, 0, 0, time.Local) }
	}
	m := NewMood()
	m.now = at(14)
	if m.Face() != faceHappy {
		t.Fatalf("default should be happy, got %q", m.Face())
	}
	m.SetUnread(1)
	if m.Face() != faceSparkly {
		t.Fatalf("new unread should be sparkly, got %q", m.Face())
	}
	m.SetCPU(95)
	if m.Face() != faceSparkly {
		t.Fatal("one high CPU sample should not fluster")
	}
	m.SetCPU(95)
	if m.Face() != faceFlustered {
		t.Fatalf("sustained CPU should fluster, got %q", m.Face())
	}
	m.SetOnline(false)
	if m.Face() != faceLost {
		t.Fatalf("offline should be lost, got %q", m.Face())
	}
	m.now = at(2)
	if m.Face() != faceSleepy {
		t.Fatalf("late night should be sleepy, got %q", m.Face())
	}
	m.now = at(14)
	m.SetOnline(true)
	m.SetCPU(10)
	m.now = func() time.Time { return at(14)().Add(sparklyDuration + time.Minute) }
	if m.Face() != faceHappy {
		t.Fatalf("sparkly should wear off, got %q", m.Face())
	}
}

func TestKaomojiOnlyUpdatesOnChange(t *testing.T) {
	m := NewMood()
	m.now = func() time.Time { return time.Date(2026, 10, 13, 14, 0, 0, 0, time.Local) }
	k := NewKaomoji(m, Colors{})
	b := &Batch{}
	k.Poll(context.Background())(b)
	if !slices.Contains(b.Args(), "label="+faceHappy) {
		t.Fatalf("unexpected face: %v", b.Args())
	}
	if k.Poll(context.Background()) != nil {
		t.Fatal("unchanged face should not update")
	}
}

const netstatOut = `Name       Mtu   Network       Address            Ipkts Ierrs     Ibytes    Opkts Oerrs     Obytes  Coll
en1        1500  <Link#17>   ea:05:5a:1c:de:78 1436070176     0 1909630884971 297217818     0 62728020035     0
en1        1500  fe80::181d: fe80:11::181d:b4b 1436070176     - 1909630884971 297217818     - 62728020035     -
`

const netstatNoAddress = `Name       Mtu   Network       Address            Ipkts Ierrs     Ibytes    Opkts Oerrs     Obytes  Coll
utun3      1380  <Link#25>                       100     0       5000      200     0       7000     0
`

func TestParseNetstat(t *testing.T) {
	got, err := parseNetstat(netstatOut, "en1")
	if err != nil {
		t.Fatal(err)
	}
	if got.in != 1909630884971 || got.out != 62728020035 {
		t.Fatalf("unexpected counters: %+v", got)
	}

	got, err = parseNetstat(netstatNoAddress, "utun3")
	if err != nil {
		t.Fatal(err)
	}
	if got.in != 5000 || got.out != 7000 {
		t.Fatalf("unexpected counters without address: %+v", got)
	}

	if _, err := parseNetstat(netstatOut, "en0"); err == nil {
		t.Fatal("expected error for missing interface")
	}
}

func TestNetworkRates(t *testing.T) {
	now := time.Unix(0, 0)
	r := newFakeRunner(map[string]string{
		"route":    "   route to: default\n  interface: en1\n",
		"ipconfig": "  InterfaceType : WiFi\n",
		"netstat":  "Name Mtu Network Address Ipkts Ierrs Ibytes Opkts Oerrs Obytes Coll\nen1 1500 <Link#17> aa 1 0 1000 1 0 2000 0\n",
	})
	n := NewNetwork(r, Colors{}, NewMood())
	n.now = func() time.Time { return now }

	b := &Batch{}
	n.Poll(context.Background())(b)
	if !slices.Contains(b.Args(), "icon="+iconWifi) {
		t.Fatalf("expected wifi icon: %v", b.Args())
	}

	r.outputs["netstat"] = "Name Mtu Network Address Ipkts Ierrs Ibytes Opkts Oerrs Obytes Coll\nen1 1500 <Link#17> aa 1 0 3001000 1 0 2500 0\n"
	now = now.Add(3 * time.Second)
	b = &Batch{}
	n.Poll(context.Background())(b)
	if !slices.Contains(b.Args(), "label=↓1.0M ↑167B") {
		t.Fatalf("unexpected rate label: %v", b.Args())
	}
}

func TestNetworkOffline(t *testing.T) {
	mood := NewMood()
	n := NewNetwork(newFakeRunner(map[string]string{}), Colors{}, mood)
	b := &Batch{}
	n.Poll(context.Background())(b)
	if !slices.Contains(b.Args(), "label=offline") {
		t.Fatalf("expected offline label: %v", b.Args())
	}
	mood.now = func() time.Time { return time.Date(2026, 10, 13, 14, 0, 0, 0, time.Local) }
	if mood.Face() != faceLost {
		t.Fatalf("offline network should make the face lost, got %q", mood.Face())
	}
}

func TestFormatRate(t *testing.T) {
	cases := []struct {
		bytes uint64
		secs  float64
		want  string
	}{
		{0, 1, "0B"},
		{999, 1, "999B"},
		{1500, 1, "1.5K"},
		{25_000_000, 1, "25M"},
		{100, 0, "0B"},
	}
	for _, c := range cases {
		if got := formatRate(c.bytes, c.secs); got != c.want {
			t.Errorf("formatRate(%d, %v) = %q, want %q", c.bytes, c.secs, got, c.want)
		}
	}
}

const vmStatOut = `Mach Virtual Memory Statistics: (page size of 16384 bytes)
Pages free:                                   129074.
Pages active:                                 682848.
Pages wired down:                             248726.
Pages purgeable:                               26770.
"Translation faults":                    30970711826.
Anonymous pages:                              972086.
Pages occupied by compressor:                 320731.
`

func TestParseVMStatUsed(t *testing.T) {
	got, err := parseVMStatUsed(vmStatOut)
	if err != nil {
		t.Fatal(err)
	}
	want := uint64(972086-26770+248726+320731) * 16384
	if got != want {
		t.Fatalf("parseVMStatUsed = %d, want %d", got, want)
	}

	if _, err := parseVMStatUsed("garbage"); err == nil {
		t.Fatal("expected error without page size")
	}
}

func TestParseCPU(t *testing.T) {
	got, err := parseCPU("  12.0\n  4.0\n 0.0\n", 4)
	if err != nil {
		t.Fatal(err)
	}
	if got != 4 {
		t.Fatalf("parseCPU = %v, want 4", got)
	}
	if got, _ := parseCPU("500\n", 2); got != 100 {
		t.Fatalf("parseCPU should cap at 100, got %v", got)
	}
}

func TestPickActiveSession(t *testing.T) {
	now := time.Now()
	sessions := []utenaSession{
		{Name: "old", IsAttached: true, LastUsedAt: now.Add(-time.Hour)},
		{Name: "detached", IsAttached: false, LastUsedAt: now},
		{Name: "new", IsAttached: true, LastUsedAt: now.Add(-time.Minute)},
	}
	if got := pickActiveSession(sessions); got != "new" {
		t.Fatalf("pickActiveSession = %q, want new", got)
	}
	if got := pickActiveSession(nil); got != "" {
		t.Fatalf("pickActiveSession(nil) = %q", got)
	}
}

func TestUtenaStream(t *testing.T) {
	var mu sync.Mutex
	active := "new-bg"
	activations := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/sessions":
			mu.Lock()
			name := active
			mu.Unlock()
			json.NewEncoder(w).Encode(map[string]any{"sessions": []map[string]any{
				{"name": name, "is_attached": true, "last_used_at": "2026-10-13T19:31:33.050163-07:00"},
			}})
		case "/tmux/activations/next":
			select {
			case <-activations:
			case <-r.Context().Done():
			}
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	updates := make(chan Update, 4)
	u := NewUtena(srv.URL, Colors{})
	u.settle = 10 * time.Millisecond
	go u.Stream(ctx, updates)

	for _, want := range []string{"new-bg", "sayo"} {
		select {
		case u := <-updates:
			b := &Batch{}
			u(b)
			if !slices.Contains(b.Args(), "label="+sparkleFrame(want)) || !slices.Contains(b.Args(), "drawing=on") {
				t.Fatalf("unexpected utena update: %v", b.Args())
			}
		case <-time.After(2 * time.Second):
			t.Fatal("timed out waiting for utena update")
		}
		mu.Lock()
		active = "sayo"
		mu.Unlock()
		activations <- struct{}{}
	}
}

func TestUtenaRefreshesWithoutActivation(t *testing.T) {
	var mu sync.Mutex
	attached := true
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/sessions":
			mu.Lock()
			isAttached := attached
			mu.Unlock()
			json.NewEncoder(w).Encode(map[string]any{"sessions": []map[string]any{
				{"name": "new-bg", "is_attached": isAttached, "last_used_at": "2026-10-13T19:31:33.050163-07:00"},
			}})
		case "/tmux/activations/next":
			<-r.Context().Done()
		}
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	updates := make(chan Update, 4)
	u := NewUtena(srv.URL, Colors{})
	u.pollWindow, u.settle = 50*time.Millisecond, time.Millisecond
	go u.Stream(ctx, updates)

	<-updates
	mu.Lock()
	attached = false
	mu.Unlock()
	select {
	case up := <-updates:
		b := &Batch{}
		up(b)
		if !slices.Contains(b.Args(), "drawing=off") {
			t.Fatalf("detach should hide the item: %v", b.Args())
		}
	case <-time.After(2 * time.Second):
		t.Fatal("detach was not picked up by the poll window")
	}
}

func TestUtenaHiddenWhenDown(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	updates := make(chan Update, 1)
	go NewUtena("http://127.0.0.1:1", Colors{}).Stream(ctx, updates)

	select {
	case u := <-updates:
		b := &Batch{}
		u(b)
		if !slices.Contains(b.Args(), "drawing=off") {
			t.Fatalf("expected hidden utena item: %v", b.Args())
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for utena update")
	}
}

func TestBluetoothPopup(t *testing.T) {
	r := newFakeRunner(map[string]string{
		"blueutil --paired":     `[{"address":"aa-bb","name":"AirPods","connected":true},{"address":"cc-dd","name":"Keyboard","connected":false}]`,
		"blueutil --connect":    "",
		"blueutil --disconnect": "",
	})
	bt := NewBluetooth(r, Colors{})

	b := &Batch{}
	bt.Click(context.Background(), Event{Name: bluetoothItem, Sender: "mouse.clicked"}, b)
	args := b.Args()
	for _, want := range [][]string{
		{"--remove", `/bluetooth\.device\..*/`},
		{"--add", "item", "bluetooth.device.0", "popup." + bluetoothItem},
		{"--add", "item", "bluetooth.device.1", "popup." + bluetoothItem},
		{"--subscribe", "bluetooth.device.1", "mouse.clicked"},
		{"--set", bluetoothItem, "popup.drawing=on"},
	} {
		if !containsSeq(args, want...) {
			t.Errorf("missing %v in %v", want, args)
		}
	}

	bt.Click(context.Background(), Event{Name: "bluetooth.device.1", Sender: "mouse.clicked"}, &Batch{})
	bt.Click(context.Background(), Event{Name: "bluetooth.device.0", Sender: "mouse.clicked"}, &Batch{})
	var toggles [][]string
	for _, c := range r.calls {
		if c[0] == "blueutil" && (c[1] == "--connect" || c[1] == "--disconnect") {
			toggles = append(toggles, c)
		}
	}
	want := [][]string{{"blueutil", "--connect", "cc-dd"}, {"blueutil", "--disconnect", "aa-bb"}}
	if !slices.EqualFunc(toggles, want, slices.Equal) {
		t.Fatalf("toggles = %v, want %v", toggles, want)
	}

	b = &Batch{}
	bt.Click(context.Background(), Event{Name: bluetoothItem, Sender: "mouse.exited.global"}, b)
	if !containsSeq(b.Args(), "--set", bluetoothItem, "popup.drawing=off") {
		t.Fatalf("expected popup to close: %v", b.Args())
	}

	b = &Batch{}
	bt.Click(context.Background(), Event{Name: bluetoothItem, Sender: "mouse.exited.global"}, b)
	if len(b.Args()) != 0 {
		t.Fatalf("closed popup should not be touched again: %v", b.Args())
	}
}

type recordingItem struct {
	events  []string
	handled []Event
	clicks  []Event
}

func (r *recordingItem) Setup(ctx context.Context, b *Batch) error {
	b.Add("rec", "right")
	return nil
}
func (r *recordingItem) Events() []string { return r.events }
func (r *recordingItem) Handle(ctx context.Context, ev Event, b *Batch) {
	r.handled = append(r.handled, ev)
	b.Set("rec", "label="+ev.Sender)
}
func (r *recordingItem) Prefix() string { return "rec" }
func (r *recordingItem) Click(ctx context.Context, ev Event, b *Batch) {
	r.clicks = append(r.clicks, ev)
}

func TestDaemonSetupSubscribesSinkOnce(t *testing.T) {
	r := newFakeRunner(nil)
	a := &recordingItem{events: []string{"front_app_switched", "system_woke"}}
	c := &recordingItem{events: []string{"system_woke"}}
	d := NewDaemon(Config{}, r, "emit", []Item{a, c})
	if err := d.Setup(context.Background()); err != nil {
		t.Fatal(err)
	}
	calls := r.sketchybarCalls()
	if len(calls) != 1 {
		t.Fatalf("setup should be a single sketchybar call, got %d", len(calls))
	}
	args := calls[0]
	if !containsSeq(args, "--subscribe", sinkItem, "front_app_switched", "system_woke") {
		t.Fatalf("sink should subscribe to each event once: %v", args)
	}
	if slices.Index(args, sinkItem) < slices.Index(args, "rec") {
		t.Fatalf("sink should subscribe after items register their events: %v", args)
	}
	if !slices.Contains(args, "script=emit") {
		t.Fatalf("sink should run the emit command: %v", args)
	}
}

func TestDaemonDispatchCoalescesEvents(t *testing.T) {
	a := &recordingItem{events: []string{"front_app_switched", "space_windows_change"}}
	d := NewDaemon(Config{}, newFakeRunner(nil), "emit", []Item{a})

	d.dispatch(context.Background(), []Event{
		{Name: sinkItem, Sender: "front_app_switched", Info: "kitty"},
		{Name: sinkItem, Sender: "volume_change"},
		{Name: sinkItem, Sender: "space_windows_change"},
	})
	if len(a.handled) != 1 || a.handled[0].Sender != "space_windows_change" {
		t.Fatalf("burst should be handled once with the latest event: %+v", a.handled)
	}

	d.dispatch(context.Background(), []Event{{Name: sinkItem, Sender: "volume_change"}})
	if len(a.handled) != 1 {
		t.Fatalf("unsubscribed events should be ignored: %+v", a.handled)
	}
}

func TestDaemonClicksRunOffLoop(t *testing.T) {
	a := &recordingItem{}
	d := NewDaemon(Config{}, newFakeRunner(nil), "emit", []Item{a})

	d.dispatch(context.Background(), []Event{{Name: "rec.child", Sender: "mouse.clicked"}, {Name: "other", Sender: "mouse.clicked"}})
	select {
	case <-d.updates:
	case <-time.After(2 * time.Second):
		t.Fatal("click result was not sent back to the loop")
	}
	if len(a.clicks) != 1 || a.clicks[0].Name != "rec.child" {
		t.Fatalf("unexpected clicks: %+v", a.clicks)
	}
}

func TestDaemonRenderHandlesEveryItem(t *testing.T) {
	a := &recordingItem{events: []string{"front_app_switched"}}
	c := &recordingItem{events: []string{"system_woke"}}
	d := NewDaemon(Config{}, newFakeRunner(nil), "emit", []Item{a, c})
	d.render(context.Background())
	if len(a.handled) != 1 || len(c.handled) != 1 {
		t.Fatalf("render should handle every item once: %+v %+v", a.handled, c.handled)
	}
}

func TestBatchOnClickUsesEmitCommand(t *testing.T) {
	b := (&Daemon{emit: "emit-cmd"}).newBatch()
	b.OnClick("x", "mouse.exited.global")
	want := []string{"--set", "x", "script=emit-cmd", "--subscribe", "x", "mouse.clicked", "mouse.exited.global"}
	if !slices.Equal(b.Args(), want) {
		t.Fatalf("OnClick = %v, want %v", b.Args(), want)
	}
}

func shortSocketPath(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "sb")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	return filepath.Join(dir, "s.sock")
}

func TestEmitAndServe(t *testing.T) {
	socket := shortSocketPath(t)
	l, err := Listen(socket)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	events := make(chan Event, 1)
	go Serve(ctx, l, events)

	sent := Event{Name: sinkItem, Sender: "aerospace_workspace_change", Env: map[string]string{"FOCUSED_WORKSPACE": "3"}}
	if err := Emit(socket, sent); err != nil {
		t.Fatal(err)
	}
	select {
	case got := <-events:
		if got.Sender != sent.Sender || got.Env["FOCUSED_WORKSPACE"] != "3" {
			t.Fatalf("unexpected event: %+v", got)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for event")
	}
}

func TestListenReplacesPreviousDaemon(t *testing.T) {
	socket := shortSocketPath(t)
	stale, err := net.Listen("unix", socket)
	if err != nil {
		t.Fatal(err)
	}
	defer stale.Close()

	old := exec.Command("sleep", "30")
	if err := old.Start(); err != nil {
		t.Fatal(err)
	}
	exited := make(chan struct{})
	go func() {
		old.Wait()
		close(exited)
	}()
	if err := os.WriteFile(socket+".pid", []byte(strconv.Itoa(old.Process.Pid)), 0o644); err != nil {
		t.Fatal(err)
	}

	server, err := Listen(socket)
	if err != nil {
		t.Fatalf("listen should take over: %v", err)
	}
	select {
	case <-exited:
	case <-time.After(3 * time.Second):
		old.Process.Kill()
		t.Fatal("previous daemon was not stopped")
	}

	pid, err := readPid(socket + ".pid")
	if err != nil || pid != os.Getpid() {
		t.Fatalf("pidfile should hold the new daemon, got %d (%v)", pid, err)
	}
	server.Close()
	if _, err := os.Stat(socket + ".pid"); !os.IsNotExist(err) {
		t.Fatalf("pidfile should be removed on close: %v", err)
	}
}

func TestListenIgnoresPidfileWithoutLiveSocket(t *testing.T) {
	socket := shortSocketPath(t)
	if err := os.WriteFile(socket+".pid", []byte("1"), 0o644); err != nil {
		t.Fatal(err)
	}
	server, err := Listen(socket)
	if err != nil {
		t.Fatal(err)
	}
	server.Close()
}

func TestEventFromEnv(t *testing.T) {
	t.Setenv("NAME", sinkItem)
	t.Setenv("SENDER", "aerospace_workspace_change")
	t.Setenv("FOCUSED_WORKSPACE", "2")
	ev := EventFromEnv()
	if ev.Name != sinkItem || ev.Sender != "aerospace_workspace_change" || ev.Env["FOCUSED_WORKSPACE"] != "2" {
		t.Fatalf("unexpected event: %+v", ev)
	}
}

func TestLoadConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bar.json")
	if err := os.WriteFile(path, []byte(`{"badges":["Mail","Slack"],"colors":{"island":"0xCC000000"}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(cfg.Badges, []string{"Mail", "Slack"}) || cfg.Colors.Island != "0xCC000000" {
		t.Fatalf("unexpected config: %+v", cfg)
	}
}

type failingRunner struct {
	*fakeRunner
	failures int
}

func (f *failingRunner) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	out, err := f.fakeRunner.Run(ctx, name, args...)
	if name == "sketchybar" && len(args) > 0 && args[0] == "--set" && f.failures > 0 {
		f.failures--
		return nil, fmt.Errorf("sketchybar: exit status 1: [!] Set: Item not found 'rec'")
	}
	return out, err
}

func TestDaemonReconfiguresWhenItemsGoMissing(t *testing.T) {
	r := &failingRunner{fakeRunner: newFakeRunner(nil), failures: 1}
	m := NewKaomoji(NewMood(), Colors{})
	m.last.changed("x")
	d := NewDaemon(Config{}, r, "emit", []Item{&recordingItem{events: []string{"system_woke"}}, m})

	b := d.newBatch()
	b.Set("rec", "label=x")
	d.apply(context.Background(), b)

	countSetups := func() int {
		n := 0
		for _, c := range r.sketchybarCalls() {
			if len(c) > 0 && c[0] == "--remove" {
				n++
			}
		}
		return n
	}
	if setups := countSetups(); setups != 1 {
		t.Fatalf("expected one re-setup after a missing item, got %d", setups)
	}
	if !m.last.changed("x") {
		t.Fatal("re-setup should clear cached item state")
	}

	b = d.newBatch()
	b.Set("rec", "label=y")
	r.failures = 1
	d.apply(context.Background(), b)
	if countSetups() != 1 {
		t.Fatal("re-setup should back off")
	}
}
