package deploy

import (
	"crypto/ed25519"
	"crypto/rand"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/crypto/ssh"
)

// mkPublicKey 返回一个一次性的 ed25519 ssh.PublicKey 用于测试。
func mkPublicKey(t *testing.T) ssh.PublicKey {
	t.Helper()
	pub, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("gen ed25519: %v", err)
	}
	sshKey, err := ssh.NewPublicKey(pub)
	if err != nil {
		t.Fatalf("new ssh key: %v", err)
	}
	return sshKey
}

func dummyAddr(t *testing.T) net.Addr {
	t.Helper()
	addr, _ := net.ResolveTCPAddr("tcp", "1.2.3.4:22")
	return addr
}

func TestTOFU_FirstConnectTrustsAndPersists(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "known_hosts")

	var logs []string
	logger := func(msg string) { logs = append(logs, msg) }

	cb := TrustOnFirstUseHostKeyCallback(path, logger)
	key := mkPublicKey(t)

	if err := cb("example.com:22", dummyAddr(t), key); err != nil {
		t.Fatalf("first connect should trust, got %v", err)
	}

	// 文件应已创建并包含一行指纹
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read known_hosts: %v", err)
	}
	if len(data) == 0 {
		t.Error("known_hosts should not be empty after TOFU")
	}
	if !strings.Contains(strings.Join(logs, "\n"), "首次连接") {
		t.Errorf("expected 首次连接 log, got %v", logs)
	}
}

func TestTOFU_SecondConnectWithSameKeyAccepted(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "known_hosts")
	cb := TrustOnFirstUseHostKeyCallback(path, nil)
	key := mkPublicKey(t)

	if err := cb("host.example.com:22", dummyAddr(t), key); err != nil {
		t.Fatalf("first: %v", err)
	}
	// 第二次连接相同 key 必须通过
	if err := cb("host.example.com:22", dummyAddr(t), key); err != nil {
		t.Errorf("second connect should succeed, got %v", err)
	}
}

func TestTOFU_FingerprintChangeRejected(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "known_hosts")
	cb := TrustOnFirstUseHostKeyCallback(path, nil)

	key1 := mkPublicKey(t)
	key2 := mkPublicKey(t)

	if err := cb("foo.example.com:22", dummyAddr(t), key1); err != nil {
		t.Fatalf("first: %v", err)
	}
	// 指纹变化：必须硬拒绝
	err := cb("foo.example.com:22", dummyAddr(t), key2)
	if err == nil {
		t.Fatal("expected rejection when host key changes, got nil")
	}
	if !strings.Contains(err.Error(), "主机密钥已变更") {
		t.Errorf("expected 主机密钥已变更 error, got %v", err)
	}
}

func TestTOFU_FallsBackToInMemoryWhenPathEmpty(t *testing.T) {
	cb := TrustOnFirstUseHostKeyCallback("", nil)
	key1 := mkPublicKey(t)
	key2 := mkPublicKey(t)

	if err := cb("bar.example.com:22", dummyAddr(t), key1); err != nil {
		t.Fatalf("first: %v", err)
	}
	if err := cb("bar.example.com:22", dummyAddr(t), key1); err != nil {
		t.Errorf("same key replay should succeed, got %v", err)
	}
	if err := cb("bar.example.com:22", dummyAddr(t), key2); err == nil {
		t.Error("expected in-memory TOFU to reject fingerprint change")
	}
}

// 不同 host 互不影响
func TestTOFU_DifferentHostsIsolated(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "known_hosts")
	cb := TrustOnFirstUseHostKeyCallback(path, nil)

	keyA := mkPublicKey(t)
	keyB := mkPublicKey(t)

	if err := cb("a.example.com:22", dummyAddr(t), keyA); err != nil {
		t.Fatalf("a: %v", err)
	}
	// 新的 host 即使 key 完全不同也应被 TOFU 接受
	if err := cb("b.example.com:22", dummyAddr(t), keyB); err != nil {
		t.Errorf("unrelated new host should be TOFU-trusted, got %v", err)
	}
	// 各自复连仍应通过
	if err := cb("a.example.com:22", dummyAddr(t), keyA); err != nil {
		t.Errorf("a second connect: %v", err)
	}
	if err := cb("b.example.com:22", dummyAddr(t), keyB); err != nil {
		t.Errorf("b second connect: %v", err)
	}
}
