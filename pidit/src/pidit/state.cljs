(ns pidit.state
  "Mutable state.
  (not including various pixi graphic objects)")

; width and height of screen
(def screen-w (atom (.-innerWidth js/window)))
(def screen-h (atom (.-innerHeight js/window)))

; current x- and y-offset (in top-left corner) from origin coordinates
(def x-off (atom 0))
(def y-off (atom 0))

; canvas max and min offsets (how big is it currently)
(def max-x-off (atom 0))
(def min-x-off (atom 0))
(def min-y-off (atom 0))
(def max-y-off (atom 0))

(def new-widget (js-obj "active" false))

(def widgets #js [])

; anonymous widget counter
(def widget-anon-c (atom 0))
