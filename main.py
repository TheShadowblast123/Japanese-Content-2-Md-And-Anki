from jisho_api.word import Word
from jisho_api.kanji import Kanji
import MeCab
import os
import glob
import re
import string
import json
import googletrans as gt
#api functions, get data from outside sources in order to study flashcards

def parse (sentence):
    mecab = MeCab.Tagger("-O wakati")
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
def intake_content ():
    output = {}
    new_content_path = './New Content'
    output_path = './Output'


    
    txt_files = glob.glob(os.path.join(new_content_path, "*.txt"))
    if txt_files == []:
        print(f'You have no new sources place .txt files in {new_content_path} in order to begin')
        return txt_files
    for txt_file in txt_files:
        name = os.path.basename(txt_file)
        with open(txt_file, "r", encoding='UTF8') as file:

            file_contents = file.read()
            result = re.sub(r'[A-Za-z0-9]', '', file_contents).replace(' ', '')
            if len(result) > 0:
                output[name] = result
    return output

def write_to_report ():
    
    return

def get_sentences ():
    hiragana_range, katakana_range = (0x3041, 0x3096), (0x30A0, 0x30FF)
    hiragana_chars, katakana_chars  = [chr(c) for c in range(*hiragana_range)], [chr(c) for c in range(*katakana_range)]

    japanese_chars = katakana_chars + hiragana_chars
    punctuation = list(string.punctuation)
    sources = intake_content()
    for k, v in sources.items():
        output = {}
        sentence = ''
        for c in v:
            if c not in punctuation:
                sentence += c
            else:
                if sentence:
                    if k not in output:
                        output[k] = []
                    output[k].append(sentence)
                    sentence = ''
    for k, v in sources.items():
        if k not in output:
            output[k] = []
        if sentence:
            output[k].append(sentence)
    return output

def write_to_output ():
    input = get_sentences()
    for name, sentences in input:
        for sentence in sentences:
            translation  = gt.translate(sentence, 'en', 'ja')
            full_sentence = ' '.join(parse(sentence))
            print(translation, full_sentence)
            break

    output = {}
    return output
def write_to_database ():
    
    return
write_to_output()