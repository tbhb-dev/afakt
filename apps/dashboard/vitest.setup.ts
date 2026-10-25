// React's act() needs this flag to flush updates without warning.
Reflect.set(globalThis, "IS_REACT_ACT_ENVIRONMENT", true);
