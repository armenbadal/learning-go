package main

import (
	"context"
	"io"
	"net/http"
	"time"
)

func fetch(ctx context.Context, url string, delay time.Duration) (string, error) {
	requestCtx, cancel := context.WithTimeout(ctx, delay)
	defer cancel()

	request, err := http.NewRequestWithContext(requestCtx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		if requestCtx.Err() != nil {
			return "", requestCtx.Err()
		}
		return "", err
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		if requestCtx.Err() != nil {
			return "", requestCtx.Err()
		}
		return "", err
	}

	return string(responseBody), nil
}

func request(url string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	result, err := fetch(ctx, url, 1*time.Second)
	if err != nil {
		return err
	}

	println(result)

	return nil
}

func main() {
	const exampleUrl = "https://cacm.acm.org/federal-funding-of-academic-research/the-llvm-compiler-infrastructure/"
	request(exampleUrl)
}
