package testing

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/zond/godip"
	"github.com/zond/godip/orders"
	"github.com/zond/godip/state"
)

func AssertOrderValidity(t *testing.T, validator godip.Validator, order godip.Order, nat godip.Nation, err error) {
	if gotNat, e := order.Validate(validator); e != err {
		t.Errorf("%v should validate to %v, but got %v", order, err, e)
	} else if gotNat != nat {
		t.Errorf("%v should validate with %q as issuer, but got %q", order, nat, gotNat)
	}
}

func AssertMove(t *testing.T, j *state.State, src, dst godip.Province, success bool) {
	if success {
		unit, _, ok := j.Unit(src)
		if !ok {
			t.Errorf("Should be a unit at %v", src)
		}
		j.SetOrder(src, orders.Move(src, dst))
		j.Next()
		if err, ok := j.Resolutions()[src]; ok && err != nil {
			t.Errorf("Move from %v to %v should have worked, got %v", src, dst, err)
		}
		if now, _, ok := j.Unit(src); ok && reflect.DeepEqual(now, unit) {
			t.Errorf("%v should have moved from %v", now, src)
		}
		if now, _, _ := j.Unit(dst); !reflect.DeepEqual(now, unit) {
			t.Errorf("%v should be at %v now", unit, dst)
		}
	} else {
		unit, _, _ := j.Unit(src)
		j.SetOrder(src, orders.Move(src, dst))
		j.Next()
		if err, ok := j.Resolutions()[src]; !ok || err == nil {
			t.Errorf("Move from %v to %v should not have worked", src, dst)
		}
		if now, _, _ := j.Unit(src); !reflect.DeepEqual(now, unit) {
			t.Errorf("%v should not have moved from %v", now, src)
		}
	}
}

func AssertUnit(t *testing.T, j *state.State, province godip.Province, unit godip.Unit) {
	if found, _, _ := j.Unit(province); !reflect.DeepEqual(found, unit) {
		t.Errorf("%v should be at %v now", unit, province)
	}
}

func AssertNoUnit(t *testing.T, j *state.State, province godip.Province) {
	_, _, ok := j.Unit(province)
	if ok {
		t.Errorf("There should be no unit at %v now", province)
	}
}

func AssertNoOptionToMoveTo(t *testing.T, j *state.State, nat godip.Nation, src godip.Province, dst godip.Province) {
	options := j.Phase().Options(j, nat)[src]
	if _, ok := options[godip.Move][godip.SrcProvince(src)][dst]; ok {
		t.Errorf("There should be no option for %v to move %v to %v", nat, src, dst)
	}
}

func AssertOptionToMove(t *testing.T, j *state.State, nat godip.Nation, src godip.Province, dst godip.Province) {
	options := j.Phase().Options(j, nat)[src]
	if _, ok := options[godip.Move][godip.SrcProvince(src)][dst]; !ok {
		t.Errorf("There should be an option for %v to move %v to %v", nat, src, dst)
	}
}

func hasOptHelper(opts map[string]interface{}, order []string, originalOpts map[string]interface{}, originalOrder []string) error {
	if len(order) == 0 {
		return nil
	}
	if _, found := opts[order[0]]; !found {
		b, err := json.MarshalIndent(originalOpts, "  ", "  ")
		if err != nil {
			return err
		}
		b2, err := json.MarshalIndent(opts, "  ", "  ")
		if err != nil {
			return err
		}
		return fmt.Errorf("Got no option for %+v in %s, failed at %+v in %s, wanted it!", originalOrder, b, order, b2)
	}
	return hasOptHelper(opts[order[0]].(map[string]interface{})["Next"].(map[string]interface{}), order[1:], originalOpts, originalOrder)
}

func hasOpt(opts godip.Options, order []string) error {
	b, err := json.MarshalIndent(opts, "  ", "  ")
	if err != nil {
		return err
	}
	converted := map[string]interface{}{}
	if err := json.Unmarshal(b, &converted); err != nil {
		return err
	}
	return hasOptHelper(converted, order, converted, order)
}

func AssertOpt(t *testing.T, opts godip.Options, order []string) {
	t.Run(strings.Join(order, "_"), func(t *testing.T) {
		err := hasOpt(opts, order)
		if err != nil {
			t.Error(err)
		}
	})
}

func AssertNoOpt(t *testing.T, opts godip.Options, order []string) {
	t.Run(strings.Join(order, "_"), func(t *testing.T) {
		err := hasOpt(opts, order)
		if err == nil {
			b, err := json.MarshalIndent(opts, "  ", "  ")
			if err != nil {
				t.Fatal(err)
			}
			t.Errorf("Found option for %+v in %s, didn't want it", order, b)
		}
	})
}

func AssertOwner(t *testing.T, j *state.State, supplyCenter string, owner godip.Nation) {
	nation, _, ok := j.SupplyCenter(godip.Province(supplyCenter))
	if !ok {
		t.Errorf("Province %s was not owned", supplyCenter)
	}
	if nation != owner {
		t.Errorf("Province %s was owned by %s, but expected %s", supplyCenter, nation, owner)
	}
}

func AssertNoOwner(t *testing.T, j *state.State, supplyCenter string) {
	nation, _, ok := j.SupplyCenter(godip.Province(supplyCenter))
	if ok {
		t.Errorf("Province %s was owned by %s, but expected no owner", supplyCenter, nation)
	}
}
