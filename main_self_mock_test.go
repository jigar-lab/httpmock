// main_self_mock_test.go
package main

import (
    "net/http"
    "testing"
    "time"

    "github.com/aws/aws-sdk-go/aws/credentials"
    "github.com/aws/aws-sdk-go/aws/signer/v4"
    "github.com/stretchr/testify/assert"
)

func TestPresignedURL(t *testing.T) {
    creds := credentials.NewStaticCredentials(
        "AKIAIOSFODNN7EXAMPLE",
        "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
        "",
    )
    signer := v4.NewSigner(creds)

    req, err := http.NewRequest("GET", "https://test-bucket.s3.us-west-2.amazonaws.com/test-object", nil)
    assert.NoError(t, err)

    // Sign the request
    _, err = signer.Sign(req, nil, "s3", "us-west-2", time.Now())
    assert.NoError(t, err)

    // Verify the signed headers contain the expected AWS signature components
    assert.Contains(t, req.Header.Get("Authorization"), "AWS4-HMAC-SHA256")
    assert.NotEmpty(t, req.Header.Get("X-Amz-Date"))
    assert.NotEmpty(t, req.Header.Get("X-Amz-Content-Sha256"))

    // Print headers for inspection
    t.Logf("Authorization: %s", req.Header.Get("Authorization"))
    t.Logf("X-Amz-Date: %s", req.Header.Get("X-Amz-Date"))
    t.Logf("X-Amz-Content-Sha256: %s", req.Header.Get("X-Amz-Content-Sha256"))
}

// If you specifically need to test presigned URLs, you can use the SignedRequest method
func TestSignedRequest(t *testing.T) {
    creds := credentials.NewStaticCredentials(
        "AKIAIOSFODNN7EXAMPLE",
        "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
        "",
    )
    signer := v4.NewSigner(creds)

    req, err := http.NewRequest("GET", "https://test-bucket.s3.us-west-2.amazonaws.com/test-object", nil)
    assert.NoError(t, err)

    // Add query parameters that you want to be signed
    query := req.URL.Query()
    query.Add("X-Amz-Expires", "900") // 15 minutes
    req.URL.RawQuery = query.Encode()

    // Sign the request
    _, err = signer.Sign(req, nil, "s3", "us-west-2", time.Now())
    assert.NoError(t, err)

    // Verify the signed request contains the expected components
    assert.Contains(t, req.Header.Get("Authorization"), "AWS4-HMAC-SHA256")
    assert.NotEmpty(t, req.Header.Get("X-Amz-Date"))
    assert.NotEmpty(t, req.Header.Get("X-Amz-Content-Sha256"))

    // Print the full URL with query parameters
    t.Logf("Signed URL: %s", req.URL.String())
}
