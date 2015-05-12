(ns pidit.js
  "Javascript interop.")

(defn !
  "Alter the fields of a javascript object. Returns the object."
  [obj & kvs]
  {:pre [(even? (count kvs))]} ; not strictly necessary, as partition will discard any trailing element
  (doseq [[k v] (partition 2 kvs)]
    (aset obj k v))
  obj)

(defn ?
  [obj k]
  (aget obj k))
