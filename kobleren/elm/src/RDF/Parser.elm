module RDF.Parser exposing (parseTriples)

import RDF.RDF exposing (..)
import Parser exposing (..)


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


parseLiteral : Parser Term
parseLiteral =
    inContext "Literal" <|
        succeed TermLiteral
            |= oneOf
                [ try parseLangLiteral
                , try parseTypedLiteral
                , parseStringLiteral
                ]


try : Parser a -> Parser a
try parser =
    delayedCommitMap always parser (succeed ())


parseStringLiteral : Parser Literal
parseStringLiteral =
    succeed Literal
        |. symbol "\""
        |= literalString
        |. symbol "\""
        |= succeed Nothing
        |= succeed xsdString


parseLangLiteral : Parser Literal
parseLangLiteral =
    succeed Literal
        |. symbol "\""
        |= literalString
        |. symbol "\""
        |. symbol "@"
        |= parseLangTag
        |= succeed rdfLangString


parseTypedLiteral : Parser Literal
parseTypedLiteral =
    succeed Literal
        |. symbol "\""
        |= literalString
        |. symbol "\""
        |. symbol "^^"
        |= succeed Nothing
        |= parseURI



-- TODO make parseAnyLiteral work


parseAnyLiteral : Parser Literal
parseAnyLiteral =
    succeed Literal
        |. symbol "\""
        |= literalString
        |. symbol "\""
        |= oneOf
            [ delayedCommit (symbol "@") <|
                parseLangTag
            , succeed Nothing
            ]
        |= succeed xsdString


blankNodeString : Parser String
blankNodeString =
    mapWithSource always <|
        ignoreWhile (\char -> char /= ' ')


uriString : Parser String
uriString =
    mapWithSource always <|
        --illegal URI characters: /[\x00-\x20<>\\"\{\}\|\^\`]/
        ignoreWhile (\char -> char /= '>')


literalString : Parser String
literalString =
    mapWithSource always <|
        ignoreWhile (\char -> char /= '"' && char /= '\n')


parseLangTag : Parser (Maybe String)
parseLangTag =
    succeed Just
        |= langString


langString : Parser String
langString =
    mapWithSource always <|
        ignoreWhile (\char -> char /= ' ' && char /= '.')


whitespace : Parser ()
whitespace =
    ignoreWhile (\char -> char == ' ' || char == '\t')


parseTriples : String -> Parser (List TriplePattern)
parseTriples ntriples =
    succeed identity
        |= zeroOrMore parseTriple
