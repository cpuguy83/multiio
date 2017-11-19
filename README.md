# multiio
Go package for treating multiple IO streams as one

### Usage

```go
  // example IO streams
  r1 := strings.NewReader("hello")
  r2 := strings.NewReader(" world")
  r3 := strings.NewReader("!!!")

  r := NewMultiReader(r1, r2, r3)
  data, _ := ioutil.ReadAll(r)   // data == "hello world!!!"
```
