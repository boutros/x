module RDF.RDF exposing (..)

--https://medium.com/elm-shorts/intro-to-records-in-elm-51bc5e933a57#.vzc4nw54i
--https://medium.com/elm-shorts/an-intro-to-constructors-in-elm-57af7a72b11e#.v3w49nxr9
-- https://github.com/RubenVerborgh/N3.js/blob/master/lib/N3Lexer.js#L50
--https://github.com/robstewart57/rdf4h/blob/master/src/Data/RDF/Types.hs
--https://github.com/TravisWhitaker/rdf/blob/master/src/Data/RDF/Internal.hs
{-



            or

            type Literal
             = Literal
                 { value : String
                 , lang : Maybe String
                 , datatype : URI
                 }


         type URI
             = URI { uri : String }


         type BlankNode
             = BlankNode { id : String }


         type Subject
             = SubjectAsURI URI
             | SubjectAsBNode BlankNode


         type Object
             = ObjectAsURI URI
             | ObjectAsBNode BlankNode
             | ObjectAsLiteral Literal


         type Triple
             = Triple
                 { subject : Subject
                 , predicate : URI
                 , object : Object
                 }

      type alias Triple =
          = Subject Predicate Object

   "<http://example.org/resource1> <http://example.org/property> <http://example.org/resource2> .
   _:anon <http://example.org/property> <http://example.org/resource2> .
   <http://example.org/resource2> <http://example.org/property> _:anon ."
   =>
       Triple
           { subject = URI "http://example.org/resource1"
           , predicate = URI "http://example.org/property"
           , object = URI "http://example.org/resource2" }
       Triple
           {
            subject = BlankNode "anon"
           }
-}


type alias Literal =
    { value : String
    , language : Maybe String
    , datatype : Term
    }


type Term
    = TermURI String
    | TermLiteral Literal
    | TermBlankNode String
    | TermVar String


type alias TriplePattern =
    { subject : Term
    , predicate : Term
    , object : Term
    }


xsdString =
    TermURI "http://www.w3.org/2001/XMLSchema#string"


rdfLangString =
    TermURI "http://www.w3.org/1999/02/22-rdf-syntax-ns#langString"
