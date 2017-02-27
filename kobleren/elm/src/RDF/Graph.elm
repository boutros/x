module RDF.Graph exposing (fromString)

import RDF.RDF exposing (..)
import Parser exposing (..)


--https://gist.github.com/jinjor/f6dd8c75d9a3e84a1b39afb4e3bee322#file-parsertrouble-elm-L119-L132
--https://groups.google.com/forum/#!topic/elm-dev/gJ-FzttCcjA
--parseGraph : Parser (List TriplePattern)
--parseLiteral : Parser Term


parseTriple : Parser TriplePattern
parseTriple =
    succeed TriplePattern
        |. whitespace
        |= parseSubject
        |. whitespace
        |= parseURI
        |. whitespace
        |= parseObject
        |. whitespace
        |. symbol "."


parseObject : Parser Term
parseObject =
    oneOf
        [ parseURI
          -- , parseLiteral
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
    succeed TermBlankNode
        |. keyword "_:"
        |= blankNodeString


parseURI : Parser Term
parseURI =
    succeed TermURI
        |. symbol "<"
        |= uriString
        |. symbol ">"



--illegal URI characters: /[\x00-\x20<>\\"\{\}\|\^\`]/


blankNodeString : Parser String
blankNodeString =
    mapWithSource always <|
        ignoreWhile (\char -> char /= ' ')


uriString : Parser String
uriString =
    mapWithSource always <|
        ignoreWhile (\char -> char /= '>')


whitespace : Parser ()
whitespace =
    ignoreWhile (\char -> (char == ' ' || char == '\t'))


parseTriples : String -> Parser (List TriplePattern)
parseTriples ntriples =
    succeed identity
        |= zeroOrMore parseTriple


fromString : String -> Result Error (List TriplePattern)
fromString ntriples =
    run (parseTriples ntriples) ntriples
