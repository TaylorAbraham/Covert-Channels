/**
 * This is a typings polyfill file. This is used strictly to polyfill typings not properly
 * provided by an npm package. You likely do not need to change anything in here unless you
 * know for sure there is an error in 3rd party typings.
 */

 // Polyfill react-bootstrap FormControl to allow e.target.value
interface EventTarget {
  value?: any;
}
