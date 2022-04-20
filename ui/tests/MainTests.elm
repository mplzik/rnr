module MainTests exposing (..)

import Expect exposing (Expectation)
import Fuzz exposing (Fuzzer, int, list, string)
import Test exposing (..)

import Html exposing (..)
import Html.Attributes exposing (..)

import Main

suite : Test
suite =
    describe "The Main module"
        [ describe "autolink" <|
                let
                    prefix = "Hello, "
                    suffix = " world!"
                    url = "http://example.com"
                    url2 = "http://example.net"
                    tests = [
                        ("doesn't change non-link message", prefix ++ suffix, [text (prefix ++ suffix)]),
                        ("converts link to a single anchor tag" , url, [a [attribute "href" url] [text url]]),
                        ("keeps text prefix of a link", prefix ++ url, [text prefix, a [attribute "href" url] [text url]]),
                        ("keeps text suffix of a link", url ++ suffix, [a [attribute "href" url] [text url], text suffix]),
                        ("converts multiple links", url ++ " " ++ url2, [a [attribute "href" url] [text url], text " ", a [attribute "href" url2] [text url2]]) ]

                in
                    List.map (\(name, message, html) -> (test name (\_ -> Expect.equal html (Main.autolink message)))) tests

                
        ]
