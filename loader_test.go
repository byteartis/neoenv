package neoenv_test

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/byteartis/neoenv"
)

type sampleConfig struct {
	Database struct {
		Host string `env:"host"`
	} `env:"database"`
	Subscribers struct {
		OrderManagement string `env:"order_management"`
	} `env:"subscribers"`
	Crons struct {
		Replay string
	}
	GracefulShutdown struct {
		Enabled bool `env:"is_enabled"`
		Seconds uint `env:"seconds"`
	}
	RootInt       int       `env:"root_int"`
	RootFloat     float64   `env:"root_float"`
	ListOfStrings []string  `env:"list_of_strings"`
	ListOfInts    []int     `env:"list_of_ints"`
	ListOfUInts   []uint    `env:"list_of_uints"`
	ListOfFloats  []float32 `env:"list_of_floats"`
}

func TestLoadPointerToNonStruct(t *testing.T) {
	t.Parallel()

	_, err := neoenv.Load[int]()
	assert.Contains(t, err.Error(), "expected pointer to struct, got pointer to *int")
}

func TestLoad(t *testing.T) {
	expectedHost := "localhost"
	expectedSusbcriberOrderManagement := "subscriber_order_management"
	expectedCronReplay := "cron-replay"
	expectedGracefulShutdownEnabled := true
	var expectedGracefulShutdownSeconds uint = 10
	expectedRootInt := -42
	expectedRootFloat := 3.14
	expectedListOfStrings := []string{"one", "two", "three"}
	expectedListOfInts := []int{1, 2, 3}
	expectedListOfFloats := []float32{1.1, 2.2, 3.3}

	t.Setenv("DATABASE__HOST", expectedHost)
	t.Setenv("SUBSCRIBERS__ORDER_MANAGEMENT", expectedSusbcriberOrderManagement)
	t.Setenv("CRONS__REPLAY", expectedCronReplay)
	t.Setenv("GRACEFUL_SHUTDOWN__IS_ENABLED", strconv.FormatBool(expectedGracefulShutdownEnabled))
	t.Setenv("GRACEFUL_SHUTDOWN__SECONDS", strconv.FormatUint(uint64(expectedGracefulShutdownSeconds), 10))
	t.Setenv("ROOT_INT", strconv.Itoa(expectedRootInt))
	t.Setenv("ROOT_FLOAT", fmt.Sprintf("%f", expectedRootFloat))
	t.Setenv("LIST_OF_STRINGS", strings.Join(expectedListOfStrings, ","))
	t.Setenv("LIST_OF_INTS", concatSlice(expectedListOfInts...))
	t.Setenv("LIST_OF_FLOATS", concatSlice(expectedListOfFloats...))

	cfg, err := neoenv.Load[sampleConfig]()
	require.NoError(t, err)

	assert.Equal(t, expectedHost, cfg.Database.Host)
	assert.Equal(t, expectedSusbcriberOrderManagement, cfg.Subscribers.OrderManagement)
	assert.Equal(t, expectedCronReplay, cfg.Crons.Replay)
	assert.Equal(t, expectedGracefulShutdownEnabled, cfg.GracefulShutdown.Enabled)
	assert.Equal(t, expectedGracefulShutdownSeconds, cfg.GracefulShutdown.Seconds)
	assert.Equal(t, expectedRootInt, cfg.RootInt)
	assert.Equal(t, expectedListOfStrings, cfg.ListOfStrings)
	assert.Equal(t, expectedListOfInts, cfg.ListOfInts)
	assert.Equal(t, expectedListOfFloats, cfg.ListOfFloats)
	assert.InEpsilon(t, expectedRootFloat, cfg.RootFloat, 0.1)
}

func concatSlice[T any](v ...T) string {
	out := make([]string, len(v))
	for i, val := range v {
		out[i] = fmt.Sprintf("%v", val)
	}
	return strings.Join(out, ",")
}

func ExampleLoad() {
	type Config struct {
		// When 'env' is omitted, the field name snake-cased is used as the key, e.g., 'newrelic_enabled'
		NewrelicEnabled bool
		// When 'env' is omitted, the field name snake-cased is used as the key, e.g., 'database'
		Database struct {
			// Nested properties are joined by '__', e.g., 'database__host'
			Host string `env:"host"`
			// Nested properties are joined by '__', e.g., 'database__host'
			Port uint `env:"port"`
		}
		Subscribers struct {
			// Nested properties are joined by '__', e.g., 'subscribers__order_management'
			OrderManagement string `env:"order_management"`
			// When 'env' is omitted, the field name snake-cased is used as the key, e.g., 'member_account'
			MemberAccount string
			// The 'env' tag can be set to any name without necessarily matching the field name in the structure
			Member string `env:"api_member"`
		} `env:"subscribers"`
		ListOfStrings []string  `env:"list_of_strings"`
		ListOfUInts   []uint    `env:"list_of_uints"`
		ListOfFloats  []float32 `env:"list_of_floats"`
	}

	os.Setenv("NEWRELIC_ENABLED", "true")
	os.Setenv("DATABASE__HOST", "localhost")
	os.Setenv("DATABASE__PORT", "5432")
	os.Setenv("SUBSCRIBERS__ORDER_MANAGEMENT", "subscriber_order_management")
	os.Setenv("SUBSCRIBERS__MEMBER_ACCOUNT", "subscriber_member_account")
	os.Setenv("SUBSCRIBERS__API_MEMBER", "subscriber_api_member")
	os.Setenv("LIST_OF_STRINGS", "one,two,three")
	os.Setenv("LIST_OF_UINTS", "1,2,3")
	os.Setenv("LIST_OF_FLOATS", "1.1,2.2,3.3")

	cfg, err := neoenv.Load[Config]()
	if err != nil {
		panic(err)
	}

	// Use the configuration
	_ = cfg
}
