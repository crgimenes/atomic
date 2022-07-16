# Example

```go
	bucket, err := db.UseBucket("bbs")
	if err != nil {
		panic(err)
	}

	err = bucket.Save("test", cfg)
	if err != nil {
		panic(err)
	}

	err = bucket.Load("test", &cfg)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%#v\n", cfg)

	l, err := bucket.List()
	for k, v := range l {
		fmt.Println(k, v)
	}
```



