open Tea.App
open Tea.Html

(* MODEL *)

let init () = "hei.."

(* UPDATE *)

type msg =
	| Noop

let update model = function
	| Noop -> model

(* VIEW *)

let view model : _ Vdom.t =
	div
		[]
		[ text model ]

let main =
	beginnerProgram {
		model = init ();
		update;
		view;
	}