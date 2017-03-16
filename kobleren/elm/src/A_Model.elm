module A_Model exposing (..)

import B_Message exposing (Route)
import C_Data as Data
import RDF.RDF as RDF
import RDF.Graph as Graph
import Dict


-- MODEL


type alias Model =
    { route : Route
    , error : String
    , query : String
    , results : Maybe Data.SearchResults
    , graph : Graph.Graph
    }



-- Extracting data from model


valueFromGraph : Model -> List RDF.TriplePattern -> String
valueFromGraph model patterns =
    let
        triples =
            Graph.query patterns model.graph
    in
        case triples of
            [] ->
                ""

            [ triple ] ->
                valueFromObject triple.object

            triple :: rest ->
                let
                    _ =
                        "expected one value, got several matching the pattern: " ++ (toString patterns)
                in
                    valueFromObject triple.object


valuesFromGraph : Model -> List RDF.TriplePattern -> List String
valuesFromGraph model patterns =
    let
        triples =
            Graph.query patterns model.graph
    in
        List.map (\triple -> valueFromObject triple.object) triples


valueFromObject : RDF.Term -> String
valueFromObject obj =
    case obj of
        RDF.TermLiteral l ->
            l.value

        RDF.TermURI uri ->
            uri

        _ ->
            ""



-- STATIC DATA


type Authority
    = Nationality
    | Audience


audiences =
    Dict.fromList
        [ ( "http://data.deichman.no/audience#adult", "Voksne" )
        , ( "http://data.deichman.no/audience#juvenile", "Barn og ungdom" )
        , ( "http://data.deichman.no/audience#ages0To2", "0-2 år" )
        , ( "http://data.deichman.no/audience#ages3To5", "3-5 år" )
        , ( "http://data.deichman.no/audience#ages6To8", "6-8 år" )
        , ( "http://data.deichman.no/audience#ages9To10", "9-10 år" )
        , ( "http://data.deichman.no/audience#ages11To12", "11-12 år" )
        , ( "http://data.deichman.no/audience#ages13To15", "13-15 år" )
        ]


nationalities =
    Dict.fromList
        [ ( "http://data.deichman.no/nationality#aborig", "Aboriginsk" )
        , ( "http://data.deichman.no/nationality#afg", "Afghanistan" )
        , ( "http://data.deichman.no/nationality#afr", "Afrikansk" )
        , ( "http://data.deichman.no/nationality#alb", "Albania" )
        , ( "http://data.deichman.no/nationality#alg", "Algerie" )
        , ( "http://data.deichman.no/nationality#am", "USA" )
        , ( "http://data.deichman.no/nationality#andor", "Andorra" )
        , ( "http://data.deichman.no/nationality#angol", "Angola" )
        , ( "http://data.deichman.no/nationality#antig", "Antigua-Barbuda" )
        , ( "http://data.deichman.no/nationality#arab", "Arabisk" )
        , ( "http://data.deichman.no/nationality#argen", "Argentina" )
        , ( "http://data.deichman.no/nationality#arm", "Armenia" )
        , ( "http://data.deichman.no/nationality#aserb", "Aserbajdsjan" )
        , ( "http://data.deichman.no/nationality#au", "Australia" )
        , ( "http://data.deichman.no/nationality#babyl", "Babylonsk" )
        , ( "http://data.deichman.no/nationality#baham", "Bahamas" )
        , ( "http://data.deichman.no/nationality#bahr", "Bahrain" )
        , ( "http://data.deichman.no/nationality#bangl", "Bangladesh" )
        , ( "http://data.deichman.no/nationality#barb", "Barbados" )
        , ( "http://data.deichman.no/nationality#belg", "Belgia" )
        , ( "http://data.deichman.no/nationality#beliz", "Belize" )
        , ( "http://data.deichman.no/nationality#benin", "Benin" )
        , ( "http://data.deichman.no/nationality#bhut", "Bhutan" )
        , ( "http://data.deichman.no/nationality#boliv", "Bolivia" )
        , ( "http://data.deichman.no/nationality#bosn", "Bosnia-Hercegovina" )
        , ( "http://data.deichman.no/nationality#botsw", "Botswana" )
        , ( "http://data.deichman.no/nationality#bras", "Brasil" )
        , ( "http://data.deichman.no/nationality#brun", "Brunei" )
        , ( "http://data.deichman.no/nationality#bulg", "Bulgaria" )
        , ( "http://data.deichman.no/nationality#burkin", "Burkina Faso" )
        , ( "http://data.deichman.no/nationality#burm", "Myanmar" )
        , ( "http://data.deichman.no/nationality#burund", "Burundi" )
        , ( "http://data.deichman.no/nationality#chil", "Chile" )
        , ( "http://data.deichman.no/nationality#colomb", "Colombia" )
        , ( "http://data.deichman.no/nationality#costaric", "Costa Rica" )
        , ( "http://data.deichman.no/nationality#cub", "Cuba" )
        , ( "http://data.deichman.no/nationality#d", "Danmark" )
        , ( "http://data.deichman.no/nationality#dominik", "Dominikanske republikk" )
        , ( "http://data.deichman.no/nationality#ecuad", "Ecuador" )
        , ( "http://data.deichman.no/nationality#egypt", "Egypt" )
        , ( "http://data.deichman.no/nationality#emiratarab", "Forente arabiske emirater" )
        , ( "http://data.deichman.no/nationality#eng", "England" )
        , ( "http://data.deichman.no/nationality#eritr", "Eritrea" )
        , ( "http://data.deichman.no/nationality#est", "Estland" )
        , ( "http://data.deichman.no/nationality#etiop", "Etiopia" )
        , ( "http://data.deichman.no/nationality#fi", "Finland" )
        , ( "http://data.deichman.no/nationality#filip", "Filippinene" )
        , ( "http://data.deichman.no/nationality#fr", "Frankrike" )
        , ( "http://data.deichman.no/nationality#fær", "Færøyene" )
        , ( "http://data.deichman.no/nationality#gabon", "Gabon" )
        , ( "http://data.deichman.no/nationality#gamb", "Gambia" )
        , ( "http://data.deichman.no/nationality#gas", "Madagaskar" )
        , ( "http://data.deichman.no/nationality#georg", "Georgia" )
        , ( "http://data.deichman.no/nationality#ghan", "Ghana" )
        , ( "http://data.deichman.no/nationality#gr", "Hellas" )
        , ( "http://data.deichman.no/nationality#grenad", "Grenada" )
        , ( "http://data.deichman.no/nationality#grønl", "Grønland" )
        , ( "http://data.deichman.no/nationality#guadel", "Guadeloupe" )
        , ( "http://data.deichman.no/nationality#guatem", "Guatemala" )
        , ( "http://data.deichman.no/nationality#guin", "Guinea" )
        , ( "http://data.deichman.no/nationality#guineab", "Guinea-Bissau" )
        , ( "http://data.deichman.no/nationality#guyan", "Guyana" )
        , ( "http://data.deichman.no/nationality#hait", "Haiti" )
        , ( "http://data.deichman.no/nationality#hond", "Honduras" )
        , ( "http://data.deichman.no/nationality#hviter", "Hviterussland" )
        , ( "http://data.deichman.no/nationality#ind", "India" )
        , ( "http://data.deichman.no/nationality#indian", "Indiansk" )
        , ( "http://data.deichman.no/nationality#indon", "Indonesia" )
        , ( "http://data.deichman.no/nationality#inuit", "Inuittisk" )
        , ( "http://data.deichman.no/nationality#ir", "Irland" )
        , ( "http://data.deichman.no/nationality#irak", "Irak" )
        , ( "http://data.deichman.no/nationality#iran", "Iran" )
        , ( "http://data.deichman.no/nationality#isl", "Island" )
        , ( "http://data.deichman.no/nationality#isr", "Israel" )
        , ( "http://data.deichman.no/nationality#it", "Italia" )
        , ( "http://data.deichman.no/nationality#ivor", "Elfenbenskysten" )
        , ( "http://data.deichman.no/nationality#jam", "Jamaica" )
        , ( "http://data.deichman.no/nationality#jap", "Japan" )
        , ( "http://data.deichman.no/nationality#jord", "Jordan" )
        , ( "http://data.deichman.no/nationality#jug", "Jugoslavia" )
        , ( "http://data.deichman.no/nationality#kamb", "Kambodsja" )
        , ( "http://data.deichman.no/nationality#kamer", "Kamerun" )
        , ( "http://data.deichman.no/nationality#kan", "Canada" )
        , ( "http://data.deichman.no/nationality#kappverd", "Kapp Verde" )
        , ( "http://data.deichman.no/nationality#kas", "Kasakhstan" )
        , ( "http://data.deichman.no/nationality#ken", "Kenya" )
        , ( "http://data.deichman.no/nationality#kin", "Kina" )
        , ( "http://data.deichman.no/nationality#kirg", "Kirgisistan" )
        , ( "http://data.deichman.no/nationality#komor", "Komorene" )
        , ( "http://data.deichman.no/nationality#kongol", "Den demokratiske republikken Kongo" )
        , ( "http://data.deichman.no/nationality#kongolbraz", "Republikken Kongo" )
        , ( "http://data.deichman.no/nationality#kor", "Korea" )
        , ( "http://data.deichman.no/nationality#kosov", "Kosovo" )
        , ( "http://data.deichman.no/nationality#kroat", "Kroatia" )
        , ( "http://data.deichman.no/nationality#kurd", "Kurdisk" )
        , ( "http://data.deichman.no/nationality#kuw", "Kuwait" )
        , ( "http://data.deichman.no/nationality#kypr", "Kypros" )
        , ( "http://data.deichman.no/nationality#lank", "Sri Lanka" )
        , ( "http://data.deichman.no/nationality#laot", "Laos" )
        , ( "http://data.deichman.no/nationality#latv", "Latvia" )
        , ( "http://data.deichman.no/nationality#lesot", "Lesotho" )
        , ( "http://data.deichman.no/nationality#liban", "Libanon" )
        , ( "http://data.deichman.no/nationality#liber", "Liberia" )
        , ( "http://data.deichman.no/nationality#liby", "Libya" )
        , ( "http://data.deichman.no/nationality#liecht", "Liechtenstein" )
        , ( "http://data.deichman.no/nationality#lit", "Litauen" )
        , ( "http://data.deichman.no/nationality#lux", "Luxembourg" )
        , ( "http://data.deichman.no/nationality#maked", "Makedonia" )
        , ( "http://data.deichman.no/nationality#malaw", "Malawi" )
        , ( "http://data.deichman.no/nationality#malay", "Malaysia" )
        , ( "http://data.deichman.no/nationality#mali", "Mali" )
        , ( "http://data.deichman.no/nationality#malt", "Malta" )
        , ( "http://data.deichman.no/nationality#maori", "Maori" )
        , ( "http://data.deichman.no/nationality#marok", "Marokko" )
        , ( "http://data.deichman.no/nationality#martinik", "Martinique" )
        , ( "http://data.deichman.no/nationality#mauret", "Mauritania" )
        , ( "http://data.deichman.no/nationality#maurit", "Mauritius" )
        , ( "http://data.deichman.no/nationality#mesop", "Mesopotamisk" )
        , ( "http://data.deichman.no/nationality#mex", "Mexico" )
        , ( "http://data.deichman.no/nationality#mold", "Moldovia" )
        , ( "http://data.deichman.no/nationality#moneg", "Monaco" )
        , ( "http://data.deichman.no/nationality#mong", "Mongolia" )
        , ( "http://data.deichman.no/nationality#montenegr", "Montenegro" )
        , ( "http://data.deichman.no/nationality#mosam", "Mosambik" )
        , ( "http://data.deichman.no/nationality#n", "Norge" )
        , ( "http://data.deichman.no/nationality#namib", "Namibia" )
        , ( "http://data.deichman.no/nationality#ned", "Nederland" )
        , ( "http://data.deichman.no/nationality#nep", "Nepal" )
        , ( "http://data.deichman.no/nationality#newzeal", "New Zealand" )
        , ( "http://data.deichman.no/nationality#nicarag", "Nicaragua" )
        , ( "http://data.deichman.no/nationality#nig", "Niger" )
        , ( "http://data.deichman.no/nationality#niger", "Nigeria" )
        , ( "http://data.deichman.no/nationality#nordir", "Nord-Irland" )
        , ( "http://data.deichman.no/nationality#nordkor", "Nord-Korea" )
        , ( "http://data.deichman.no/nationality#om", "Oman" )
        , ( "http://data.deichman.no/nationality#pak", "Pakistan" )
        , ( "http://data.deichman.no/nationality#pal", "Palestina" )
        , ( "http://data.deichman.no/nationality#panam", "Panama" )
        , ( "http://data.deichman.no/nationality#pap", "Papua Ny-Guinea" )
        , ( "http://data.deichman.no/nationality#parag", "Paraguay" )
        , ( "http://data.deichman.no/nationality#pers", "Persisk" )
        , ( "http://data.deichman.no/nationality#peru", "Peru" )
        , ( "http://data.deichman.no/nationality#pol", "Polen" )
        , ( "http://data.deichman.no/nationality#portug", "Portugal" )
        , ( "http://data.deichman.no/nationality#puert", "Puerto Rica" )
        , ( "http://data.deichman.no/nationality#qat", "Qatar" )
        , ( "http://data.deichman.no/nationality#r", "Russland" )
        , ( "http://data.deichman.no/nationality#rom", "Romersk" )
        , ( "http://data.deichman.no/nationality#rum", "Romania" )
        , ( "http://data.deichman.no/nationality#rwand", "Rwanda" )
        , ( "http://data.deichman.no/nationality#salvad", "El Salvador" )
        , ( "http://data.deichman.no/nationality#sam", "Samisk" )
        , ( "http://data.deichman.no/nationality#samoan", "Samoa" )
        , ( "http://data.deichman.no/nationality#sanktluc", "Sankt Lucia" )
        , ( "http://data.deichman.no/nationality#saudiarab", "Saudi-Arabia" )
        , ( "http://data.deichman.no/nationality#senegal", "Senegal" )
        , ( "http://data.deichman.no/nationality#serb", "Serbia" )
        , ( "http://data.deichman.no/nationality#sierral", "Sierra Leone" )
        , ( "http://data.deichman.no/nationality#sk", "Skottland" )
        , ( "http://data.deichman.no/nationality#slovak", "Slovakia" )
        , ( "http://data.deichman.no/nationality#sloven", "Slovenia" )
        , ( "http://data.deichman.no/nationality#somal", "Somalia" )
        , ( "http://data.deichman.no/nationality#sp", "Spania" )
        , ( "http://data.deichman.no/nationality#storbr", "Storbritannia" )
        , ( "http://data.deichman.no/nationality#sudan", "Sudan" )
        , ( "http://data.deichman.no/nationality#surin", "Surinam" )
        , ( "http://data.deichman.no/nationality#sv", "Sverige" )
        , ( "http://data.deichman.no/nationality#sveits", "Sveits" )
        , ( "http://data.deichman.no/nationality#swazil", "Swaziland" )
        , ( "http://data.deichman.no/nationality#syr", "Syria" )
        , ( "http://data.deichman.no/nationality#sørafr", "Sør-Afrika" )
        , ( "http://data.deichman.no/nationality#sørkor", "Sør-Korea" )
        , ( "http://data.deichman.no/nationality#sørsudan", "Sør-Sudan" )
        , ( "http://data.deichman.no/nationality#t", "Tyskland" )
        , ( "http://data.deichman.no/nationality#tadsj", "Tadsjikistan" )
        , ( "http://data.deichman.no/nationality#tahit", "Tahiti" )
        , ( "http://data.deichman.no/nationality#taiw", "Taiwan" )
        , ( "http://data.deichman.no/nationality#tanz", "Tanzania" )
        , ( "http://data.deichman.no/nationality#tchad", "Tchad" )
        , ( "http://data.deichman.no/nationality#thai", "Thailand" )
        , ( "http://data.deichman.no/nationality#tib", "Tibet" )
        , ( "http://data.deichman.no/nationality#togo", "Togo" )
        , ( "http://data.deichman.no/nationality#trinid", "Trinid og Tobago" )
        , ( "http://data.deichman.no/nationality#tsj", "Tsjekkia" )
        , ( "http://data.deichman.no/nationality#tsjet", "Tsjetsjenia" )
        , ( "http://data.deichman.no/nationality#tun", "Tunisia" )
        , ( "http://data.deichman.no/nationality#turkm", "Turkmenistan" )
        , ( "http://data.deichman.no/nationality#tyrk", "Tyrkia" )
        , ( "http://data.deichman.no/nationality#ugand", "Uganda" )
        , ( "http://data.deichman.no/nationality#ukr", "Ukraina" )
        , ( "http://data.deichman.no/nationality#ung", "Ungarn" )
        , ( "http://data.deichman.no/nationality#urug", "Uruguay" )
        , ( "http://data.deichman.no/nationality#usb", "Usbekistan" )
        , ( "http://data.deichman.no/nationality#venez", "Venezuela" )
        , ( "http://data.deichman.no/nationality#viet", "Vietnam" )
        , ( "http://data.deichman.no/nationality#wal", "Wales" )
        , ( "http://data.deichman.no/nationality#yemen", "Jemen" )
        , ( "http://data.deichman.no/nationality#zair", "Zaïre" )
        , ( "http://data.deichman.no/nationality#zamb", "Zambia" )
        , ( "http://data.deichman.no/nationality#zimb", "Zimbabwe" )
        , ( "http://data.deichman.no/nationality#øst", "Østerrike" )
        ]


labelFor : String -> Authority -> String
labelFor uri authority =
    let
        dict =
            case authority of
                Nationality ->
                    nationalities

                Audience ->
                    audiences
    in
        case Dict.get uri dict of
            Just label ->
                label

            Nothing ->
                "missing label for: " ++ uri


allValuesFor : Authority -> List ( String, String )
allValuesFor authority =
    let
        dict =
            case authority of
                Nationality ->
                    nationalities

                Audience ->
                    audiences
    in
        Dict.toList dict
