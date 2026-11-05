package grid

import (
	"errors"
)

var (
	// ErrUnknownMovement is wrapped by the panic value of a movement outside the constants,
	// MovementNone included, where a method has no answer for it, and by the error of
	// UnmarshalJSON reading square cells with no movement.
	ErrUnknownMovement = errors.New("grid: unknown movement")
	// ErrNonFinite is wrapped by the panic value of an origin or bounds with a NaN or infinite
	// coordinate.
	ErrNonFinite = errors.New("grid: non-finite coordinate")
	// ErrInvalidCellScale is wrapped by the panic value of a cell scale that is not positive and
	// finite on both axes, and by the error of UnmarshalJSON reading one.
	ErrInvalidCellScale = errors.New("grid: cell scale not positive and finite")
	// ErrInvalidCost is wrapped by the panic value of a step cost that is negative or NaN.
	ErrInvalidCost = errors.New("grid: negative or NaN cost")
)
