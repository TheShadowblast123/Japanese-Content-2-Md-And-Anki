from jisho_api.word import Word
from jisho_api.kanji import Kanji
import MeCab
import os
import glob
import re
import string
import googletrans as gt
import time
from concurrent.futures import ThreadPoolExecutor

def parse_sentence(sentence):
    mecab = MeCab.Tagger("-O wakati")
    return mecab.parse(sentence).split()
def kanji_data(kanji):
    request = Kanji.request(kanji)
    try:
        data = request.data
        main_readings = data.main_readings
        output = {
            'kanji_' : kanji,
            'keyword' : data.main_meanings[0],
            'readings' : [main_readings.kun, main_readings.on],
            'strokes' : data.strokes,
            'radicals' : data.radical.parts,
        }
        return output
    except:
        return
def word_data(word):
    request = Word.request(word)
    try:
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
    except:
        return
def intake_content ():
    output = {}
    new_content_path = './New Content'
    output_path = './Output'

    txt_files = glob.glob(os.path.join(new_content_path, "*.txt"))
    if txt_files == []:
        print(f'You have no new sources place .txt files in {new_content_path} in order to begin')
        return
    for txt_file in txt_files:
        name = os.path.basename(txt_file)
        with open(txt_file, "r", encoding='UTF8') as file:
            file_contents = file.read()
            result = re.sub(r'[A-Za-z0-9]', '', file_contents).replace(' ', '')
            if len(result) > 0:
                output[name] = result
    return output
def get_sentences ():
    punctuation = list(string.punctuation +'\n')
    sources = intake_content()
    for name, content in sources.items():
        output = {}
        sentence = ''
        for c in content:
            if c not in punctuation:
                sentence += c
            else:
                if sentence:
                    if name not in output:
                        output[name] = []
                    output[name].append(sentence)
                    sentence = ''
    for name, content in sources.items():
        if name not in output:
            output[name] = []
        if sentence:
            output[name].append(sentence)
    return output
def write_to_output ():
    input = get_sentences()
    for name, sentences in input.items():
        unique_sentences = list(set(sentences))
        tag_name = os.path.splitext(name)[0]
        with ThreadPoolExecutor() as executor:
            futures = [executor.submit(process_sentence, sentence, tag_name) for sentence in unique_sentences]

            # Wait for all tasks to complete
            for future in futures:
               future.result()
    return
def process_sentence(sentence, tag_name):
    parsed = parse_sentence(sentence)
    formatted_sentence = [f'[[{w}]]' for w in parsed]
    translation = gt.translate(sentence, 'en', 'ja')
    markdown_sentence = f'\nSTARTI [Basic] {formatted_sentence} Back: {translation} Tags: {tag_name}  ENDI\n'
    write_notes(markdown_sentence, parsed, tag_name)
    return
def write_notes(markdown_sentence, parsed_sentence, name):
    hiragana_range, katakana_range, punctuation_range_one, punctuation_range_two = (0x3041, 0x3096), (0x30A0, 0x30FF), (0xFF01, 0xFF5E), (0x3000, 0x303F)
    hiragana_chars, katakana_chars, punctuation_range_one, punctuation_range_two = [chr(c) for c in range(*hiragana_range)], [chr(c) for c in range(*katakana_range)], [chr(c) for c in range(*punctuation_range_one)], [chr(c) for c in range(*punctuation_range_two)]

    kana_punctuation_chars = katakana_chars + hiragana_chars + punctuation_range_one + punctuation_range_two
    path_to_sentence = 'Notes\Japanese Notes\Sentences.md'
    
    with open(path_to_sentence, "r+", encoding="utf-8") as file:
        existing_content = file.read()
        if markdown_sentence not in existing_content:
            file.seek(0, 2)
            file.write(markdown_sentence)
    #sentence has ben written
    for w in parsed_sentence:
        add_word_data(w, name)
        if w[0] not in kana_punctuation_chars:
            kanji_list = ''.join(c for c in w if c not in kana_punctuation_chars)
            for item in kanji_list:
                k = kanji_data(item)
                add_kanji_data(k, w, name)
        
    return
def add_word_data(w, source):
    path_to_words = 'Notes\Japanese Notes\Words.md'
    path_to_report = 'Output\\report.txt'
    try:
        word_data = word_data(w)
    except:
        return
    if word_data is not None:
        write = ''
        with open(path_to_words, "r+", encoding="utf-8") as file:
            existing_content = file.read()
            if w not in existing_content:
                write = f'\nSTART\n Basic\n {word_data["word"]}\n Back: {word_data["definitions"]}\n {word_data["reading"]}\n Tags: {source} \n END\n'
                file.seek(0, 2)
                file.write(write)
                return
            else:
                start_pos = existing_content.find(w)
                tags_pos = existing_content.rfind("Tags:", 0, start_pos)
                tags_end_pos = existing_content.find("\n", tags_pos)
                existing_tags = existing_content[tags_pos:tags_end_pos]

                write = f"{existing_tags} {source}"

                file.seek(tags_pos)
                file.write(write)
                return
    else:
        with open(path_to_report, "a", encoding="utf-8") as file:
            file.write(f"couldn't resolve issue with the folliwing, in {source}.txt, Look it up: {w}\n")
        return
def add_kanji_data(k, w, source):
    path_to_kanji = 'Notes\Japanese Notes\Kanji.md'
    path_to_report = 'Output\\report.txt'

    if k is not None:
        write = ''
        with open(path_to_kanji, "r+", encoding="utf-8") as file:
            existing_content = file.read()
            if k["kanji_"] not in existing_content:
                write = f'\nSTART\n Basic\n {k["kanji_"]}, {k["strokes"]}\n Back: {k["keyword"]}\n {k["readings"]}\n {k["radicals"]}\n Tags: {w} \n END\n'
                file.seek(0, 2)
                file.write(write)
                return
            else:
                start_pos = existing_content.find(k)
                tags_pos = existing_content.rfind("Tags:", 0, start_pos)
                tags_end_pos = existing_content.find("\n", tags_pos)
                existing_tags = existing_content[tags_pos:tags_end_pos]

                write = f"{existing_tags} {w}"

                file.seek(tags_pos)
                file.write(write)
                return
    else:
        with open(path_to_report, "a", encoding="utf-8") as file:
            file.write(f"couldn't resolve issue with the folliwing, in {source}.txt, Look it up: {k['kanji_']} from {w}\n")
        return
start_time = time.time()
if __name__ == "__main__":
    write_to_output()

end_time = time.time()
elapsed_time = end_time - start_time
print(f"Elapsed time: {elapsed_time} seconds")