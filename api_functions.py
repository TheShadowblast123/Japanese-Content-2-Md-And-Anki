from jisho_api.word import Word
from jisho_api.kanji import Kanji
import MeCab
import sys
def parse (sentence):
    mecab = MeCab.Tagger("-Owakat")
    return mecab.parse(sentence).split()
def kanji (kanji):
    request = Kanji.request(kanji)
    if request.meta.status == 200:

        data = request.data
        main_readings = data.main_readings
        output = {
            'kanji' : kanji,
            'keyword' : data.main_meanings[0],
            'readings' : [main_readings.kun, main_readings.on],
            'strokes' : data.strokes,
            'radicals' : data.radical.parts,
        }
        return output
    else:
        return request.meta.status
def word (word):
    

    request = Word.request(word)
    if request.meta.status == 200:
        defintions = []
        data = request.data[0]
        for defintion in data.senses:
            if defintion.parts_of_speech == ['Wikipedia definition']:
                break
            defintions.append(defintion.english_definitions)
        output = {
            'word' : word,
            'definitions' : defintions,
            'reading' : data.japanese[0].reading
        }
        return output
    else:
        return request.meta.status
