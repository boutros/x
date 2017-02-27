module Views.Authority.Person exposing (view)

import Html exposing (Html, div, text, p, main_, nav, h2, fieldset)
import Html.Attributes exposing (class)
import A_Model exposing (..)
import B_Message exposing (..)
import Views.Inputs as Inputs
import RDF.RDF as RDF
import RDF.Query as Query
import RDF.Ontology as Ontology


view : Model -> String -> Html Msg
view model uri =
    let
        subject =
            (RDF.TermURI uri)
    in
        div []
            [ nav []
                [ h2 [] [ text "Person" ]
                , div [] [ text "uri" ]
                ]
            , main_
                []
                [ div
                    [ class "resource-edit" ]
                    [ fieldset []
                        [ (Inputs.singleString (Query.matchOne subject Ontology.name) "Navn")
                        , (Inputs.singleNumber (Query.matchOne subject Ontology.birthYear) "Fødselsår" 4)
                        , (Inputs.singleNumber (Query.matchOne subject Ontology.deathYear) "Dødsår" 4)
                        , (Inputs.singleSearchSelect "Nasjonalitet" [ "Norge", "Sverige", "Danmark", "Senegal" ])
                        , (Inputs.singleString [] "Kjønn")
                        , (Inputs.singleText "Forklarende tilføyelse" 3)
                        , (Inputs.singleString [] "Alternativt navn")
                        , (Inputs.singleNumber (Query.matchOne subject Ontology.number) "Nummer" 4)
                        ]
                    ]
                ]
            ]
