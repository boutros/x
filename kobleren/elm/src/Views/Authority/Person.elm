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
                        [ (Inputs.singleString model (Query.matchOne subject Ontology.name) "Navn")
                        , (Inputs.singleNumber model (Query.matchOne subject Ontology.birthYear) "Fødselsår" 4)
                        , (Inputs.singleNumber model (Query.matchOne subject Ontology.deathYear) "Dødsår" 4)
                        , (Inputs.multiSearchSelect
                            model
                            (Query.matchOne subject Ontology.nationality)
                            "Nasjonalitet"
                            Nationality
                          )
                        , (Inputs.singleString model [] "Kjønn")
                        , (Inputs.singleText "Forklarende tilføyelse" 3)
                        , (Inputs.singleString model [] "Alternativt navn")
                        , (Inputs.singleNumber model (Query.matchOne subject Ontology.number) "Nummer" 4)
                        ]
                    ]
                ]
            ]
