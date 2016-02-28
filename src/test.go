import (
    "testing"
    "net/http"
)

func main() {

}

testCreatePaste(t *testing.T) {
    p := "test paste"

    req, err := http.NewRequest("POST", "localhost/save")
}
