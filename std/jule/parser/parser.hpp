#ifndef __JULE_STD_JULE_PARSER
#define __JULE_STD_JULE_PARSER

#include "../../../api/jule.hpp"

template<typename Vector, typename Item>
jule::Slice<Item> __jule_parser_vector_as_slice(Vector vec) noexcept;

template<typename Vector, typename Item>
jule::Slice<Item> __jule_parser_vector_as_slice(Vector vec) noexcept {
	jule::Slice<Item> slice;
	if (vec._method_len() == 0)
		return slice;

	slice._len = vec._method_len();
	slice._cap = vec._method_cap();
	slice.data.alloc = *reinterpret_cast<Item**>(vec._field__heap);
	slice._slice = &slice.data.alloc[0];
	return slice;
}

#endif // __JULE_STD_JULE_PARSER
