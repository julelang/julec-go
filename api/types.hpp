#ifndef __JULE_TYPES_HPP
#define __JULE_TYPES_HPP

#include <stddef.h>

#include "platform.hpp"

namespace jule {

#ifdef ARCH_32BIT
    typedef unsigned long int Uint;
    typedef signed long int Int;
    typedef unsigned long int Uintptr;
#else
    typedef unsigned long long int Uint;
    typedef signed long long int Int;
    typedef unsigned long long int Uintptr;
#endif

    typedef signed char I8;
    typedef signed short int I16;
    typedef signed long int I32;
    typedef signed long long int I64;
    typedef unsigned char U8;
    typedef unsigned short int U16;
    typedef unsigned long int U32;
    typedef unsigned long long int U64;
    typedef float F32;
    typedef double F64;
    typedef bool Bool;

    constexpr decltype(nullptr) nil{ nullptr };

    constexpr jule::F32 MAX_F32{ 0x1p127 * (1 + (1 - 0x1p-23)) };
    constexpr jule::F32 MIN_F32{ -0x1p127 * (1 + (1 - 0x1p-23)) };
    constexpr jule::F64 MAX_F64{ 0x1p1023 * (1 + (1 - 0x1p-52)) };
    constexpr jule::F64 MIN_F64{ -0x1p1023 * (1 + (1 - 0x1p-52)) };
    constexpr jule::I64 MAX_I64{ 9223372036854775807LL };
    constexpr jule::I64 MIN_I64{ -9223372036854775807 - 1 };
    constexpr jule::U64 MAX_U64{ 18446744073709551615LLU };

} // namespace jule

#endif // ifndef __JULE_TYPES_HPP
