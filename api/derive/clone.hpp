// Copyright 2023 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULE_DERIVE_CLONE_HPP
#define __JULE_DERIVE_CLONE_HPP

#include "../types.hpp"
#include "../str.hpp"
#include "../array.hpp"
#include "../slice.hpp"
#include "../map.hpp"
#include "../ref.hpp"
#include "../trait.hpp"
#include "../fn.hpp"

namespace jule {

    char clone(const char &x) noexcept;
    signed char clone(const signed char &x) noexcept;
    unsigned char clone(const unsigned char &x) noexcept;
    char *clone(char *x) noexcept;
    const char *clone(const char *x) noexcept;
    jule::Int clone(const jule::Int &x) noexcept;
    jule::Uint clone(const jule::Uint &x) noexcept;
    jule::Bool clone(const jule::Bool &x) noexcept;
    jule::Str clone(const jule::Str &x) noexcept;
    template<typename Item> jule::Slice<Item> clone(const jule::Slice<Item> &s) noexcept;
    template<typename Item, const jule::Uint N> jule::Array<Item, N> clone(const jule::Array<Item, N> &arr) noexcept;
    template<typename Key, typename Value> jule::Map<Key, Value> clone(const jule::Map<Key, Value> &m) noexcept;
    template<typename T> jule::Ref<T> clone(const jule::Ref<T> &r) noexcept;
    template<typename T> jule::Trait<T> clone(const jule::Trait<T> &t) noexcept;
    template<typename T> jule::Fn<T> clone(const jule::Fn<T> &fn) noexcept;
    template<typename T> T *clone(T *ptr) noexcept;
    template<typename T> const T *clone(const T *ptr) noexcept;
    template<typename T> T clone(const T &t) noexcept;

    char clone(const char &x) noexcept { return x; }
    signed char clone(const signed char &x) noexcept { return x; }
    unsigned char clone(const unsigned char &x) noexcept { return x; }
    char *clone(char *x) noexcept { return x; }
    const char *clone(const char *x) noexcept { return x; }
    jule::Int clone(const jule::Int &x) noexcept { return x; }
    jule::Uint clone(const jule::Uint &x) noexcept { return x; }
    jule::Bool clone(const jule::Bool &x) noexcept { return x; }
    jule::Str clone(const jule::Str &x) noexcept { return x; }

    template<typename Item>
    jule::Slice<Item> clone(const jule::Slice<Item> &s) noexcept {
        jule::Slice<Item> s_clone(s.len());
        for (int i{ 0 }; i < s.len(); ++i)
            s_clone.operator[](i) = jule::clone(s.operator[](i));
        return s_clone;
    }

    template<typename Item, const jule::Uint N>
    jule::Array<Item, N> clone(const jule::Array<Item, N> &arr) noexcept {
        jule::Array<Item, N> arr_clone{};
        for (int i{ 0 }; i < arr.len(); ++i)
            arr_clone.operator[](i) = jule::clone(arr.operator[](i));
        return arr_clone;
    }

    template<typename Key, typename Value>
    jule::Map<Key, Value> clone(const jule::Map<Key, Value> &m) noexcept {
        jule::Map<Key, Value> m_clone;
        for (const auto &pair: m)
            m_clone[jule::clone(pair.first)] = jule::clone(pair.second);
        return m_clone;
    }

    template<typename T>
    jule::Ref<T> clone(const jule::Ref<T> &r) noexcept {
        if (!r.real())
            return r;

        jule::Ref<T> r_clone{ jule::Ref<T>::make(jule::clone(r.operator T())) };
        return r_clone;
    }

    template<typename T>
    jule::Trait<T> clone(const jule::Trait<T> &t) noexcept {
        jule::Trait<T> t_clone{ t };
        t_clone.data = jule::clone(t_clone.data);
        return t;
    }

    template<typename T>
    jule::Fn<T> clone(const jule::Fn<T> &fn) noexcept
    { return fn; }

    template<typename T>
    T *clone(T *ptr) noexcept
    { return ptr; }

    template<typename T>
    const T *clone(const T *ptr) noexcept
    { return ptr; }

    template<typename T>
    T clone(const T &t) noexcept
    { return t.clone(); }

}; // namespace jule

#endif // ifndef __JULE_DERIVE_CLONE_HPP
