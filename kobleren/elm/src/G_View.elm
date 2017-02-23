module G_View exposing (view)

import Html exposing (Html, div, text)
import A_Model exposing (..)
import B_Message exposing (..)
import Views.Home as Home
import Views.Authority.Person as Person


view : Model -> Html Msg
view model =
    div []
        [ page model ]


page : Model -> Html Msg
page model =
    case model.route of
        HomeRoute ->
            Home.view model

        EditAuthorityRoute uri type_ ->
            Person.view model

        NotFoundRoute ->
            notFoundView model


notFoundView : Model -> Html never
notFoundView _ =
    div []
        [ text "not found" ]
