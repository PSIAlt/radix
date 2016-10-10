#include "sort.h"

#define GEN_ARRAY(T, K);;\
GEN_SORT(T, K);;\
;;\
type STRUCT(Array) struct {;;\
	data SLICE(T);;\
};;\
;;\
func (a STRUCT(Array)) Has(x K) bool {;;\
	DO_SEARCH(a.data, x, i, ok);;\
	return ok;;\
};;\
;;\
func (a STRUCT(Array)) Get(x K) T {;;\
	DO_SEARCH(a.data, x, i, ok);;\
	if !ok {;;\
		return nil;;\
	};;\
	return a.data[i];;\
};;\
;;\
func (a STRUCT(Array)) Upsert(x T) (cp STRUCT(Array), prev T) {;;\
	var with SLICE(T);;\
	DO_SEARCH(a.data, ID(x), i, has);;\
	if has {;;\
		with = make(SLICE(T), len(a.data));;\
		copy(with, a.data);;\
		a.data[i], prev = x, a.data[i];;\
	} else {;;\
		with = make(SLICE(T), len(a.data)+1);;\
		copy(with[:i], a.data[:i]);;\
		copy(with[i+1:], a.data[i:]);;\
		with[i] = x;;\
	};;\
	return STRUCT(Array){with}, prev;;\
};;\
;;\
func (a STRUCT(Array)) Delete(x K) (cp STRUCT(Array), prev T) {;;\
	DO_SEARCH(a.data, x, i, has);;\
	if !has {;;\
		return a, nil;;\
	};;\
	without := make(SLICE(T), len(a.data)-1);;\
	copy(without[:i], a.data[:i]);;\
	copy(without[i:], a.data[i+1:]);;\
	return STRUCT(Array){without}, a.data[i];;\
};;\
;;\
func (a STRUCT(Array)) Ascend(cb func(x T) bool) bool {;;\
	for _, x := range a.data {;;\
		if !cb(x) {;;\
			return false;;\
		};;\
	};;\
	return true;;\
};;\
;;\
func (a STRUCT(Array)) AscendRange(x, y K, cb func(x T) bool) bool {\
	DO_SEARCH_SHORT(a.data, x, i);;\
	DO_SEARCH_SHORT(a.data, y, j);;\
	for ; i < len(a.data) && i <= j; i++ {;;\
		if !cb(a.data[i]) {;;\
			return false;;\
		};;\
	};;\
	return true;;\
};;\
;;\
func (a STRUCT(Array)) Len() int {;;\
	return len(a.data);;\
};;\
