module Tests exposing (..)

import Test exposing (..)
import NTriples exposing (ntriples)


all : Test
all =
    describe "Kobleren"
        [ ntriples ]
