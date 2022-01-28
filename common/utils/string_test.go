package utils

import (
	"strings"
	"testing"
)

func isEqual(t *testing.T, excepted int, v int) {
	if v != excepted {
		t.Fatalf("%d != %d", excepted, v)
	}
}

func TestTemperatureStrToInt(t *testing.T) {
	r, _ := TemperatureStrToInt("123")
	isEqual(t, 123, r)
	r, _ = TemperatureStrToInt("456C")
	isEqual(t, 456, r)
}

func TestSizeStrToMiBInt(t *testing.T) {
	r, _ := SizeStrToMiBInt("123")
	isEqual(t, 0, r)
	r, _ = SizeStrToMiBInt("123KiB")
	isEqual(t, 0, r)
	r, _ = SizeStrToMiBInt("123MiB")
	isEqual(t, 123, r)
	r, _ = SizeStrToMiBInt("123GiB")
	isEqual(t, 123*1024, r)
	r, _ = SizeStrToMiBInt("123TiB")
	isEqual(t, 123*1024*1024, r)
	r, _ = SizeStrToMiBInt("123KB")
	isEqual(t, 123*1000/1024/1024, r)
	r, _ = SizeStrToMiBInt("123MB")
	isEqual(t, 123*1000*1000/1024/1024, r)
	r, _ = SizeStrToMiBInt("123GB")
	isEqual(t, 123*1000*1000*1000/1024/1024, r)
	r, _ = SizeStrToMiBInt("123TB")
	isEqual(t, 123*1000*1000*1000*1000/1024/1024, r)
}

func TestSplitToIntList(t *testing.T) {
	r, err := SplitToIntList("123")
	if err != nil {
		t.Fatalf("split error: %v", err)
	}
	if len(r) != 1 || r[0] != 123 {
		t.Fatal("r[0] != 123")
	}
	r, err = SplitToIntList("123, 456,789")
	if err != nil || len(r) != 3 || r[0] != 123 || r[1] != 456 || r[2] != 789 {
		t.Fatalf("split error: %v", r)
	}
	_, err = SplitToIntList("123, 456,789a")
	if err == nil {
		t.Fatal("error not found in here!")
	}
}

func TestPartition(t *testing.T) {
	a, b, c := Partition("http://xixi.com", "://")
	if a != "http" {
		t.Fatalf("a is %s", a)
	}
	if b != "://" {
		t.Fatalf("b is %s", b)
	}
	if c != "xixi.com" {
		t.Fatalf("c is %s", c)
	}
	a, b, c = Partition("http://xixi.com", "a")
	if a != "http://xixi.com" {
		t.Fatalf("a is %s", a)
	}
	if b != "" {
		t.Fatalf("b is %s", b)
	}
	if c != "" {
		t.Fatalf("c is %s", c)
	}
}

func TestRenderTemplate(t *testing.T) {
	tmpl := `  upstream printservice-grpc {
        server  {{ .COMMON_IP }}:10157;
        keepalive 300;
    }`
	dicts := map[string]string{
		"COMMON_IP": "127.0.0.1",
	}
	val, err := RenderTemplate(dicts, tmpl)
	if err != nil {
		t.Fatalf("failed to render tmpl, err %+v", err)
	}
	if strings.Contains(val, "COMMON_IP") {
		t.Fatalf("dont render tmpl, %s", val)
	}
}
