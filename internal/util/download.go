package util

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"

	"github.com/replicate/replicate-go"
	"golang.org/x/sync/errgroup"
)

func DownloadPrediction(ctx context.Context, prediction replicate.Prediction, dir string) error {
	if prediction.ID == "" {
		return fmt.Errorf("prediction ID is empty")
	}

	if prediction.Status != replicate.Succeeded {
		return fmt.Errorf("prediction is not finished")
	}

	if prediction.Output == nil {
		return fmt.Errorf("prediction output is empty")
	}

	if dir == "" {
		return fmt.Errorf("directory is empty")
	}

	err := os.MkdirAll(dir, 0o755)
	if err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if reflect.TypeOf(prediction.Output).Kind() == reflect.Slice {
		v := reflect.ValueOf(prediction.Output)
		strings := make([]string, v.Len())
		for i := 0; i < v.Len(); i++ {
			strings[i] = v.Index(i).Interface().(string)
		}

		g, _ := errgroup.WithContext(ctx)

		for _, str := range strings {
			u, err := url.ParseRequestURI(str)
			if err != nil {
				break
			}

			req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
			if err != nil {
				break
			}

			g.Go(func() error {
				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					return fmt.Errorf("failed to download file %v: %w", u, err)
				}
				defer resp.Body.Close()

				filename := filepath.Base(u.Path)
				file, err := os.Create(filepath.Join(dir, filename))
				if err != nil {
					return fmt.Errorf("failed to create file %s: %w", filename, err)
				}

				_, err = io.Copy(file, resp.Body)
				if err != nil {
					return fmt.Errorf("failed to write file %s: %w", filename, err)
				}

				return nil
			})
		}

		return g.Wait()
	}

	bytes, err := json.Marshal(prediction.Output)
	if err != nil {
		return fmt.Errorf("failed to marshal prediction output: %w", err)
	}

	return os.WriteFile(filepath.Join(dir, "output.json"), bytes, 0o644)
}
