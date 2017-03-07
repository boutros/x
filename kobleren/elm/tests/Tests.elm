module Tests exposing (..)

import Test exposing (..)
import NTriples exposing (ntriples)
import Graph exposing (graph)


all : Test
all =
    describe "Kobleren"
        [ ntriples
        , graph
        ]
