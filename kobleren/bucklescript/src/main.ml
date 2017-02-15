open Tea.App
open Tea.Html
module Cmd = Tea.Cmd
module Sub = Tea.Sub
module Http = Tea.Http

(* MODEL *)

type searchHit =
	{ id : string
	; label : string
	; authorityType : string
	; abstract : string
	}

type searchResults =
	{ took : int
	; totalHits : int
	; hits : searchHit list
	}

type model =
	{ error : string
	; results : searchResults option
	}

let emptyModel =
	{ error = ""
	; results = None
	}

let init () =
	emptyModel, Cmd.none

(* MESSAGES *)

(* TODO msg.ml *)

type msg =
	| Search of string
	| GetResults of Result * Http.Error * searchResults

(* TODO decoders.ml *)

(* HTTP *)

let doSearch query =
	let url = String.concat "" [ "http://localhost:8008/dummysearch?q="; query]
	in Http.send GetResults (Http.get url decodeResults)

(* UPDATE *)

let update model = function
	| GetResults -> model, Cmd.none


(* VIEW *)

let view model : _ Vdom.t =
	div
		[]
		[ text model.error ]

let main =
	standardProgram
		{ init
		; update
		; view
		; subscriptions = (fun _model -> Sub.none)
		}