package timebox

import "time"

// now is swapped out in tests to control the current time.
var now = time.Now
