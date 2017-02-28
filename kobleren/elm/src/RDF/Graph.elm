module RDF.Graph exposing (fromString)

import RDF.RDF exposing (..)
import Parser exposing (..)


--https://gist.github.com/jinjor/f6dd8c75d9a3e84a1b39afb4e3bee322#file-parsertrouble-elm-L119-L132
--https://groups.google.com/forum/#!topic/elm-dev/gJ-FzttCcjA
--parseGraph : Parser (List TriplePattern)


parseTriple : Parser TriplePattern
parseTriple =
    succeed TriplePattern
        |. whitespace
        |= parseSubject
        |. whitespace
        |= parsePredicate
        |. whitespace
        |= parseObject
        |. whitespace
        |. symbol "."


parseObject : Parser Term
parseObject =
    inContext "object" <|
        oneOf
            [ parseURI
            , parseLiteral
            , parseBlankNode
            ]


parseSubject : Parser Term
parseSubject =
    inContext "subject" <|
        oneOf
            [ parseURI
            , parseBlankNode
            ]


parsePredicate : Parser Term
parsePredicate =
    inContext "predicate" <|
        parseURI


parseBlankNode : Parser Term
parseBlankNode =
    inContext "blank node" <|
        succeed TermBlankNode
            |. keyword "_:"
            |= blankNodeString


parseURI : Parser Term
parseURI =
    inContext "URI" <|
        succeed TermURI
            |. symbol "<"
            |= uriString
            |. symbol ">"



--illegal URI characters: /[\x00-\x20<>\\"\{\}\|\^\`]/


parseLiteral : Parser Term
parseLiteral =
    inContext "Literal" <|
        succeed TermLiteral
            |= parseLit



-- TODO howto avoid this extra function parseLit, should be included in above parseLiteral


parseLit : Parser Literal
parseLit =
    succeed Literal
        |. symbol "\""
        |= literalString
        |= succeed Nothing
        |= succeed xsdString
        |. symbol "\""


blankNodeString : Parser String
blankNodeString =
    mapWithSource always <|
        ignoreWhile (\char -> char /= ' ')


uriString : Parser String
uriString =
    mapWithSource always <|
        ignoreWhile (\char -> char /= '>')


literalString : Parser String
literalString =
    mapWithSource always <|
        ignoreWhile (\char -> char /= '"' && char /= '\n')


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
