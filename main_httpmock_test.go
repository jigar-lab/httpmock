// main_httpmock_test.go
package main

import (
    "fmt"
    "net/http"
    "testing"
    "time"

    "io/ioutil"
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/credentials"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/s3"
    "github.com/jarcoal/httpmock"
    "github.com/stretchr/testify/assert"
)

func TestSimpleS3WithMock1(t *testing.T) {
        // Enable httpmock
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// Mock S3 GetObject API call
	httpmock.RegisterResponder("GET", "https://s3.us-west-2.amazonaws.com/",
		func(req *http.Request) (*http.Response, error) {
			// Check if this is a GetObject request
			if req.URL.Path == "/my-bucket/my-file.txt" {
				resp := httpmock.NewBytesResponse(200, []byte("This is the mocked content of my-file.txt"))
				resp.Header.Set("Content-Type", "text/plain")
				resp.Header.Set("Content-Length", fmt.Sprintf("%d", len("This is the mocked content of my-file.txt")))
				resp.Header.Set("Last-Modified", "Wed, 21 Oct 2015 07:28:00 GMT")
				resp.Header.Set("ETag", "\"d41d8cd98f00b204e9800998ecf8427e\"")
				return resp, nil
			}

			// If not a GetObject request, return a 404
			return httpmock.NewStringResponse(404, "Not Found"), nil
		})

	// Create a new AWS session
	sess, err := session.NewSession(&aws.Config{
		Region:     aws.String("us-west-2"),
		DisableSSL: aws.Bool(true),
		Credentials: credentials.NewStaticCredentials(
          		"AKIAIOSFODNN7EXAMPLE",
            		"wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
            		"",
        	),
	})
	if err != nil {
		t.Fatalf("Error creating session: %v", err)
		return
	}

	// Create S3 service client
	svc := s3.New(sess)

	// Specify the bucket and item key
	bucket := "my-bucket"
	item := "my-file.txt"

	// Get the item from S3 (this will use the mocked response)
	result, err := svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(item),
	})
	if err != nil {
		t.Fatalf("Error getting object: %v", err)
		return
	}
	defer result.Body.Close()

	// Read the S3 object content
	content, err := ioutil.ReadAll(result.Body)
	if err != nil {
		t.Fatalf("Error reading content: %v", err)
		return
	}

	t.Logf("Content of %s/%s: %s\n", bucket, item, string(content))

	// Print stats
	t.Logf("Calls made: %d\n", httpmock.GetTotalCallCount())
}	

func TestSimpleS3WithMock(t *testing.T) {
    // Enable httpmock
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// Mock S3 GET request
	httpmock.RegisterResponder("GET", "https://s3.amazonaws.com/my-bucket/my-file.txt",
		func(req *http.Request) (*http.Response, error) {
			resp := httpmock.NewStringResponse(200, "This is the mocked content of my-file.txt")
			resp.Header.Set("Content-Type", "text/plain")
			return resp, nil
		})

	// Create a new AWS session
	sess, err := session.NewSession(&aws.Config{
		Region:     aws.String("us-west-2"),
		DisableSSL: aws.Bool(true), // This is important for httpmock to work
		Credentials: credentials.NewStaticCredentials(
          		"AKIAIOSFODNN7EXAMPLE",
            		"wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
            		"",
        	),
	})
	if err != nil {
		t.Logf("Error creating session: %v", err)
		return
	}

	// Create S3 service client
	svc := s3.New(sess)

	// Specify the bucket and item key
	bucket := "my-bucket"
	item := "my-file.txt"

	// Get the item from S3 (this will use the mocked response)
	result, err := svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(item),
	})
	if err != nil {
		t.Logf("Error getting object: %v", err)
		return
	}
	defer result.Body.Close()

	// Read the S3 object content
	content, err := ioutil.ReadAll(result.Body)
	if err != nil {
		t.Logf("Error reading content: %v", err)
		return
	}

	t.Logf("Content of %s/%s: %s\n", bucket, item, string(content))
	
	// Print stats
	t.Logf("Calls made: %d\n", httpmock.GetTotalCallCount())
}

func TestS3PreSignedURLWithMock(t *testing.T) {
    // Enable httpmock
    httpmock.Activate()
    defer httpmock.DeactivateAndReset()

    // Create session with mock credentials
    sess := session.Must(session.NewSession(&aws.Config{
        Region: aws.String("us-west-2"),
	DisableSSL: aws.Bool(true),
        Credentials: credentials.NewStaticCredentials(
            "AKIAIOSFODNN7EXAMPLE",
            "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
            "",
        ),
    }))

    // Create S3 client
    svc := s3.New(sess)

    // Register mock response for GetObject
    httpmock.RegisterResponder("GET", `=~^https://test-bucket\.s3\.us-west-2\.amazonaws\.com/test-object.*`,
        func(req *http.Request) (*http.Response, error) {
            // Verify request has required presigned URL components
            query := req.URL.Query()
            requiredParams := []string{
                "X-Amz-Algorithm",
                "X-Amz-Credential",
                "X-Amz-Date",
                "X-Amz-Expires",
                "X-Amz-SignedHeaders",
                "X-Amz-Signature",
            }

            for _, param := range requiredParams {
                if query.Get(param) == "" {
                    return httpmock.NewStringResponse(400, fmt.Sprintf("Missing required parameter: %s", param)), nil
                }
            }

            return httpmock.NewStringResponse(200, "mock object content"), nil
        },
    )

    // Create GetObject request
    input := &s3.GetObjectInput{
        Bucket: aws.String("test-bucket"),
        Key:    aws.String("test-object"),
    }
    req, _ := svc.GetObjectRequest(input)

    // Generate presigned URL
    urlStr, err := req.Presign(15 * time.Minute)
    if err != nil {
        t.Fatalf("failed to presign request: %v", err)
    }

    // Verify the URL contains required components
    assert.Contains(t, urlStr, "X-Amz-Algorithm=AWS4-HMAC-SHA256")
    assert.Contains(t, urlStr, "X-Amz-Credential=")
    assert.Contains(t, urlStr, "X-Amz-Date=")
    assert.Contains(t, urlStr, "X-Amz-Expires=")
    assert.Contains(t, urlStr, "X-Amz-SignedHeaders=")
    assert.Contains(t, urlStr, "X-Amz-Signature=")

    // Log the generated URL
    t.Logf("Presigned URL: %s", urlStr)

    // Try to use the presigned URL
    resp, err := http.Get(urlStr)
    if err != nil {
        t.Fatalf("failed to make request: %v", err)
    }
    defer resp.Body.Close()

    assert.Equal(t, 200, resp.StatusCode)

    // Verify mock was called
    assert.Equal(t, 1, httpmock.GetTotalCallCount())
}

// Test with expired URL
func TestS3PreSignedURLExpired(t *testing.T) {
    httpmock.Activate()
    defer httpmock.DeactivateAndReset()

    sess := session.Must(session.NewSession(&aws.Config{
        Region: aws.String("us-west-2"),
        Credentials: credentials.NewStaticCredentials(
            "AKIAIOSFODNN7EXAMPLE",
            "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
            "",
        ),
    }))

    svc := s3.New(sess)

    // Register mock response for expired URL
    httpmock.RegisterResponder("GET", `=~^https://test-bucket\.s3\.us-west-2\.amazonaws\.com/test-object.*`,
        func(req *http.Request) (*http.Response, error) {
            return httpmock.NewStringResponse(403, "Request has expired"), nil
        },
    )

    // Create GetObject request
    input := &s3.GetObjectInput{
        Bucket: aws.String("test-bucket"),
        Key:    aws.String("test-object"),
    }
    req, _ := svc.GetObjectRequest(input)

    // Generate URL that expires in 1 second
    urlStr, err := req.Presign(1 * time.Second)
    if err != nil {
        t.Fatalf("failed to presign request: %v", err)
    }

    // Wait for URL to expire
    time.Sleep(2 * time.Second)

    // Try to use the expired URL
    resp, err := http.Get(urlStr)
    if err != nil {
        t.Fatalf("failed to make request: %v", err)
    }
    defer resp.Body.Close()

    assert.Equal(t, 403, resp.StatusCode)
}

// Test with custom headers
func TestS3PreSignedURLWithCustomHeaders(t *testing.T) {
    httpmock.Activate()
    defer httpmock.DeactivateAndReset()

    sess := session.Must(session.NewSession(&aws.Config{
        Region: aws.String("us-west-2"),
        Credentials: credentials.NewStaticCredentials(
            "AKIAIOSFODNN7EXAMPLE",
            "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
            "",
        ),
    }))

    svc := s3.New(sess)

    // Register mock response
    httpmock.RegisterResponder("GET", `=~^https://test-bucket\.s3\.us-west-2\.amazonaws\.com/test-object.*`,
        func(req *http.Request) (*http.Response, error) {
            query := req.URL.Query()
            if query.Get("response-content-disposition") == "" {
                return httpmock.NewStringResponse(400, "Missing content disposition"), nil
            }
            return httpmock.NewStringResponse(200, "mock object content"), nil
        },
    )

    input := &s3.GetObjectInput{
        Bucket:                     aws.String("test-bucket"),
        Key:                        aws.String("test-object"),
        ResponseContentDisposition: aws.String("attachment; filename=test.txt"),
    }
    req, _ := svc.GetObjectRequest(input)

    urlStr, err := req.Presign(15 * time.Minute)
    if err != nil {
        t.Fatalf("failed to presign request: %v", err)
    }

    resp, err := http.Get(urlStr)
    if err != nil {
        t.Fatalf("failed to make request: %v", err)
    }
    defer resp.Body.Close()

    assert.Equal(t, 200, resp.StatusCode)
}
