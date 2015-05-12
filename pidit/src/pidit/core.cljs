(ns pidit.core
  (:require [pidit.pixi :as pi]
            [pidit.dom :as dom]
            [pidit.config :as cfg]
            [pidit.js :refer [! ?]]
            [pidit.state :as state]))

; Setup -----------------------------------------------------------------------

(enable-console-print!)
(println "hei")

; Models ----------------------------------------------------------------------

(deftype Widget [x y w h name])

; Canvas & graphics  ----------------------------------------------------------

(def texture (pi/texture-from-image cfg/bg-image))
(def tile (pi/tiling-sprite texture [@state/screen-w @state/screen-h]))

(def renderer
  (pi/renderer-auto [@state/screen-w @state/screen-h] {:backgroundColor 0xFFFFFF :antialias true}))

; infinite canvas
(def canvas (pi/container))

; widgets graphics object
(def g-widgets (pi/graphics))

; temproary graphics object (for short-lived graphics)
(def g-temp (pi/graphics))

; Event handlers -------------------------------------------------------------

(defn on-mouse-down
  [event]
  (let [x (.. event -data -originalEvent -offsetX)
        y (.. event -data -originalEvent -offsetY)]
    (! state/new-widget "x" x "y" y "width" 0 "height" 0 "active" true)))

(defn on-mouse-up
  [event]
  (let [x (? state/new-widget "x")
        y (? state/new-widget "y")
        w (? state/new-widget "width")
        h (? state/new-widget "height")]
    (swap! state/widget-anon-c inc)
    (.push state/widgets (Widget. x y w h (str "anonymous [" @state/widget-anon-c "]")))
    (! state/new-widget "active" false)
    (-> g-widgets (pi/line-style 1 0x000000 1)
          (pi/begin-fill 0xffffff 1)
          (pi/draw-rect [x y] w h)
          (pi/end-fill))))

(defn on-mouse-move
  [event]
  (if (? state/new-widget "active")
    (let [x (? state/new-widget "x")
          y (? state/new-widget "y")
          mouse-x (.. event -data -originalEvent -offsetX)
          mouse-y (.. event -data -originalEvent -offsetY)]
      (! state/new-widget "width" (- mouse-x x))
      (! state/new-widget "height" (- mouse-y y))
      (-> g-temp
          (pi/clear)
          (pi/line-style 2 0x000000 1)
          (pi/begin-fill 0xffff99 0.5)
          (pi/draw-rect [x y]
                        (? state/new-widget "width") (? state/new-widget "height"))
          (pi/end-fill)))
    ; else clear current
    (-> g-temp (pi/clear))))

(defn resize
  []
  (reset! state/screen-w (.-innerWidth js/window))
  (reset! state/screen-h (.-innerHeight js/window))
  (.resize renderer @state/screen-w @state/screen-h)
  (set! (.-width tile) @state/screen-w)
  (set! (.-height tile) @state/screen-h))

(set! (.-onresize js/window) resize)

(resize)

(pi/set-interactive! canvas true)
(pi/on-event! canvas :mousedown on-mouse-down)
(pi/on-event! canvas [:mouseup :mouseupoutside] on-mouse-up)
(pi/on-event! canvas :mousemove on-mouse-move)

(dom/set-styles! (.-view renderer)
                 "display" "block"
                 "position" "absolute"
                 "top" "0px"
                 "left" "0px")

(dom/add-or-replace-node! "app" (.-view renderer))

(pi/add! canvas tile)
(pi/add! canvas g-widgets)
(pi/add! canvas g-temp)

; Main loop -------------------------------------------------------------------
; 
; actions:
; - move around in canvas (=adjust x-off & y-off)
;   click outside of any widgets and move mouse
; - add widget to canvas
; - focus widget (=adjust x-off & y-off so that widget is centered)
; - move widget in canvas
; - remove widget from canvas

(defn animate
  []
  (js/requestAnimationFrame animate)
  (pi/render! renderer canvas))

(animate)

