module Views.Authority.Corporation exposing (view)

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
                [ h2 [] [ text "Organisasjon" ]
                , div [] [ text "uri" ]
                ]
            , main_
                []
                [ div
                    [ class "resource-edit" ]
                    [ fieldset []
                        [ (Inputs.singleString model (Query.matchOne subject Ontology.name) "Navn")
                        , (Inputs.multiString model (Query.matchOne subject Ontology.altLabel) "Altarnativt navn")
                        , (Inputs.singleString model (Query.matchOne subject Ontology.subdivision) "Underavdeling")
                        , (Inputs.singleString model (Query.matchOne subject Ontology.specification) "Forklarende tilføyelse")
                        , (Inputs.singleString model (Query.matchOne subject Ontology.place) "Sted")
                        , (Inputs.multiSearchSelect
                            model
                            (Query.matchOne subject Ontology.nationality)
                            "Nasjonalitet"
                            Nationality
                          )
                        ]
                    ]
                ]
            ]
