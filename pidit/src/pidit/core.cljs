(ns pidit.core)

(enable-console-print!)

(println "heisk")

(defn main []
  (let [c (.. js/document (createElement "DIV"))]
    (aset c "innerHTML" "<p>i was dynamically created</p>")
    (.. js/document (getElementById "app") (appendChild c))))

