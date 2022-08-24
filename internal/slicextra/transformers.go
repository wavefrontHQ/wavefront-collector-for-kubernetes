package slicextra

// FilterMapInPlace filters and transforms the given elements in-place.
// The provided function should return false when you want to filter the element.
// It nils out any left-over elements in the slice.
func FilterMapInPlace[E any](f func(E) (E, bool), es []E) []E {
	newEs := es[:0]
	for _, e := range es {
		newE, include := f(e)
		if include {
			newEs = append(newEs, newE)
		}
	}
	for i := range es[len(newEs):] {
		var empty E
		es[len(newEs)+i] = empty
	}
	return newEs
}

// MapInPlace transforms the given elements in-place.
func MapInPlace[E any](transform func(E) E, es []E) []E {
	for i, e := range es {
		es[i] = transform(e)
	}
	return es
}
