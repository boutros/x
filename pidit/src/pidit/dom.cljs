(ns pidit.dom)

(defn by-id
  "Returns the dom element with the given id."
  [id]
  (.getElementById js/document id))

(defn remove-children!
  "Removes all children of dom element."
  [elem]
  (while (.hasChildNodes elem)
    (.removeChild elem (aget (.-children elem) 0))))

(defn add-or-replace-node!
  "Adds dom node to element with the given id, or replaces it if not empty."
  [id node]
  (let [elem (by-id id)]
    (remove-children! elem)
    (.appendChild elem node)))

(defn set-styles!
  [obj & kvs]
  {:pre [(even? (count kvs))]}
  (doseq [[k v] (partition 2 kvs)]
    (aset obj "style" k v))
  obj)