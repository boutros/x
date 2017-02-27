module RDF.Graph exposing (..)

import RDF.RDF exposing (..)
import Parser exposing (..)


--https://gist.github.com/jinjor/f6dd8c75d9a3e84a1b39afb4e3bee322#file-parsertrouble-elm-L119-L132
--https://groups.google.com/forum/#!topic/elm-dev/gJ-FzttCcjA
--parseGraph : Parser (List TriplePattern)
--parseLiteral : Parser Term


parseTriple : Parser TriplePattern
parseTriple =
    succed identity
        |. whitespace
        |= parseSubject
        |. whitespace
        |= parseURI
        |. whitespace
        |. parseObject
        |. whitespace
        |. symbol "."


parseObject : Parser Term
parseObject =
    oneOf
        [ parseURI
        , parseLiteral
        , parseBlankNode
        ]


parseSubject : Parser Term
parseSubject =
    oneOf
        [ parseURI
        , parseBlankNode
        ]


parseBlankNode : Parser Term
parseBlankNode =
    succed (TermBlankNode identity)
        |. keyword "_:"
        |= string


parseURI : Parser Term
parseURI =
    succed (TermURI identity)
        |. symbol "<"
        |= uriString
        |. symbol ">"



--illegal URI characters: /[\x00-\x20<>\\"\{\}\|\^\`]/


uriString : Parser String
uriString =
    mapWithSource always <|
        ignoreWhile (\char -> char /= '>')


whitespace : Parser ()
whitespace =
    ignoreWhile (\char -> (char == ' ' || char == '\t'))


fromString : String -> Result (List TriplePattern)
fromString ntriples =
    []
