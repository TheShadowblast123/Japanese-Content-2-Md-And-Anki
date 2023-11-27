from jisho_api.word import Word
from jisho_api.kanji import Kanji
import MeCab
import os
import glob
import re
import googletrans as gt
import string
from concurrent.futures import ThreadPoolExecutor
old_kanji = []
old_words = []
global content_md, kanji_md, sentences_md, words_md, content_path, kanji_path, sentences_path, words_path
global current_name
current_name = ''
content_md = "Notes\Japanese Notes\Content.md"
kanji_md = "Notes\Japanese Notes\Kanji.md"
sentences_md = "Notes\Japanese Notes\Sentences.md"
words_md = "Notes\Japanese Notes\Words.md"
content_path = "Notes\Japanese Notes\Content"
kanji_path = "Notes\Japanese Notes\Kanji"
sentences_path = "Notes\Japanese Notes\Sentences"
words_path = "Notes\Japanese Notes\Words"
kanji_range_1, kanji_range_2, kanji_range_3 = (0x3400, 0x4DB5),(0x4E00,0x9FCB), (0xF900, 0xFA6A)
kanji_set_1, kanji_set_2, kanji_set_3 = [chr(c) for c in range(*kanji_range_1)], [chr(c) for c in range(*kanji_range_2)], [chr(c) for c in range(*kanji_range_3)]
kanji_set = kanji_set_1 + kanji_set_2 + kanji_set_3
def parser(item) -> list[str]:
    mecab = MeCab.Tagger("-O wakati")
    return mecab.parse(item).split()
def check_title(title, test):
    return test == title.replace(' ', '').replace('[', '').replace(']', '')
def intake_content ():
    output = {}
    new_content_path = "./New Content"


    txt_files = glob.glob(os.path.join(new_content_path, "*.txt"))
    if txt_files == []:
        print(f'You have no new sources place .txt files in {new_content_path} in order to begin')
        return
    for txt_file in txt_files:
        name = replace_spaces(os.path.basename(txt_file)).strip('.txt')
        with open(txt_file, "r", encoding='UTF8') as file:
            lines = file.readlines()
            file.close()
        blob = ''
        for line in lines:
            if line not in blob:
                temp = re.sub(r'[A-Za-z0-9]', '', line).replace(' ', '')
                if len(temp) > 0:
                    blob += temp
            continue
        output[name] = blob
        this_content_md = content_path + f'{name}.md'
        with open(this_content_md, 'w', encoding='utf8') as file:
            file.writelines(lines)
            file.close()
    return output
def get_sentences ():
    punctuation = ['\n', '.', '?', '!', ' 〪', '。', ' 〭', '！', '．', '？']
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
                    if sentence not in output[name]:
                        output[name].append(sentence)
                    sentence = ''
    for name, content in sources.items():
        if name not in output:
            output[name] = []
        if sentence:
            output[name].append(sentence)
   
    return output

def sentence_to_word_string(sentence):
    word_punctuation = string.punctuation + r'！＂”“＃＄％＆＇（）＊＋，－．／：；＜＝＞？＠［＼］＾＿｀｛｜｝～、。〃〄々〆〇〈〉《》「」『』【】〒〓〔〕〖〗〘〙〚〛〜〝〞〟〠〡〢〣〤〥〦〧〨〩〪〭〮〯〫〬〰〱〲〳〴〵〶〷〸〹〺〻〼〽〾｟｠｡｢｣､･〿'
    pattern = f'[{re.escape(word_punctuation)}]'
    temp = str(parser(sentence))
    words_string = re.sub(pattern, '', temp)
    words_string = words_string.split()
    temp_array = [f'[[{word}]]' for word in words_string]
    words_string = ' '.join(temp_array)

    return words_string
def word_to_kanji_string(word):
    temp = [ f'[[{x}]]' if x in kanji_set else x for x in word]
    kanji = ''.join(temp)
    return kanji
def replace_spaces (tag):
    return tag.replace(' ', '_')
def append_content(name):
    with open(content_md, 'a+') as file:
        x = f'{name}\n'
        file.write(x)
        file.close()
        #we assume the content is new no matter what
    return
def append_sentence(sentence):
    with open(sentences_md, 'a+',  encoding="utf-8") as file:
        temp = file.readlines()
        for t in temp:
            if check_title(t, sentence):
                file.close()
                return True       
        x = f'{sentence}\n'
        file.write(x)
        file.close()
        return False
def append_word(word):
    with open(words_md, 'a+', encoding="utf-8") as file:
        temp = file.readlines()
        for t in temp:
            if check_title(t, word):
                file.close()
                return True     
        x = f'{word}\n'
        file.write(x)
        file.close()
        return False
def append_kanji(kanji):
    with open(kanji_md, 'a+', encoding="utf-8") as file:
        temp = file.readlines()
        for t in temp:
            if check_title(t, kanji):
                file.close()
                return True    
        x = f'{kanji}\n'
        file.write(x)
        file.close()
        return False

def edit_tags(root_path, tag, edit_list):
    for item in edit_list:
        tag = replace_spaces(tag)
        with open (f'{root_path}\{item}.md', 'w+') as file:
            lines = file.readlines()
            for line in lines:
                if 'Tags: ' in line:
                    lines[lines.index(line)] = line + f'[[{tag}]] '
                    break
                continue
            file.writelines(lines)
            file.close()
    return
def write_to_kanji(l):
    edit_kanji_list = []
    for k in l:
        if append_kanji(k):
            edit_kanji_list.append(l.pop(k))
    return edit_kanji_list
def write_to_words(l):
    edit_words_list = []
    for k in l:
        if append_word(k):
            edit_words_list.append(l.pop(k))
    return edit_words_list
def write_to_sentences(l):
    edit_sentences_list = []
    for k in l:
        if append_sentence(k):
            edit_sentences_list.append(l.pop(k))
    return edit_sentences_list

def sentence_card(data):
    temp = [] 
    sentence = data["sentence"]
    temp.append('TARGET DECK: Sentences')
    temp.append('START')
    temp.append('Basic')
    temp.append(f'{sentence_to_word_string(sentence)}')
    temp.append('Back: ' + f'{data["translation"]}')
    temp.append(f'Tags: [[{current_name}]]\n')
    temp.append('END')
    output ='\n'.join(temp)
    write_card(output, f'{sentences_path}\{sentence}.md')
def word_card(data : dict): 
    temp = []
    word = data['word']
    temp.append('TARGET DECK: Words')
    temp.append('START')
    temp.append('Basic')
    temp.append(f'{word_to_kanji_string(data["word"])}')
    temp.append('Back: ' + f'{data["definitions"]}')
    temp.append(f'{data["reading"]}')
    temp.append(f'Tags: [[{current_name}]]\n')
    temp.append('END')
    output ='\n'.join(temp)
    write_card(output, f'{words_path}\{word}.md')
def kanji_card(data : dict):
    temp = []
    kanji = data['kanji_']
    temp.append('TARGET DECK: Kanji')
    temp.append('START')
    temp.append('Basic')
    temp.append(f'{kanji}, {data["strokes"]}')
    temp.append('Back: ' + f'{data["keyword"]}')
    temp.append(f'{data["readings"]}')
    temp.append(f'{data["radicals"]}')
    temp.append(f'Tags: [[{current_name}]]\n')
    temp.append('END')
    output ='\n'.join(temp)
    write_card(output, f'{kanji_path}\{kanji}.md')
def kanji_data(kanji):
    request = Kanji.request(kanji)
    try:
        data = request.data
        main_readings = data.main_readings
        output = {
            'kanji_' : kanji,
            'keyword' : data.main_meanings[0],
            'readings' : (*main_readings.kun, *main_readings.on),
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
            print(f"Now would be a good time to write to report.txt or something eh? {word} is returning empty")
            return {
            'word' : f'[[word]]',
            'definitions' : [],
            'reading' : ''
        }
def sentence_data(sentence):
    output = {
        'sentence' : sentence,
        'translation' : gt.translate(sentence, 'en', 'ja'),

    }
    return output
def write_card(lines : str, path : str):
    with open(path, 'w',encoding="utf-8") as file:
        file.write(lines)
    return
def write_sentence_cards(sentences : list[str]):
    this_content_md = content_path + f'{name}.md'
    with open(this_content_md, 'a+', encoding='utf8') as file:
        for line in sentences:
            file.write(f'[{line}]' + '\n')
        file.close()
    for s in sentences:
        sentence_card(sentence_data(s))
    return

def write_word_cards(words : list[str]):
    for w in words:
        word_card(word_data(w))
    return
def write_kanji_cards(kanjis : list[str]):
    for kanji in kanjis:
        kanji_card(kanji_data(kanji))
    return

new_content = get_sentences()
for name, sentences in new_content.items():
    current_name = name
    temp = ''
    for sentence in sentences:
        temp += sentence
    words = []
    words_temp = parser(temp)
    
    for w in words_temp:
        if w not in words:
            words.append(w)

    kanji_list_temp = [k for k in kanji_set if k in temp]
    kanji_list = []
    for k in kanji_list_temp:
        if k not in kanji_list:
            kanji_list.append(k)

    #handle first parallel, write to {langpart}.md and add duplicates to a list for editing tags and remove them for the orignal list
    
    with ThreadPoolExecutor() as executor:
        future_a = executor.submit(append_content, name)
        future_b = executor.submit(write_to_sentences, sentences)
        future_c = executor.submit(write_to_words, words)
        future_d = executor.submit(write_to_kanji, kanji_list)
    
        executor.shutdown(wait=True)

        edit_sentences = future_b.result()
        edit_words = future_c.result()
        edit_kanji = future_d.result()

    with ThreadPoolExecutor() as executor:
        if len(edit_kanji) > 0:
            executor.map(edit_tags, edit_kanji) 
        if len(edit_sentences) > 0:
            executor.submit(edit_tags, edit_sentences) 
        if len(edit_words) > 0:
            executor.submit(edit_tags, edit_words) 
        
        executor.shutdown(wait=True)
    with ThreadPoolExecutor() as executor:
        executor.map(write_kanji_cards, kanji_list)
        executor.submit(write_sentence_cards, sentences)
        executor.submit(write_word_cards, words)
        executor.shutdown(wait=True)

        
