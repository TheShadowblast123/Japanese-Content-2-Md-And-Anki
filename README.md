# Japanese Content 2 Md And Anki

## WARNING CURRENT IMPLEMENTATION FEATURES A TEST ONLY DIRECTORY THAT GETS DELETED ON PROGRAM START

## RoadMap

### Path to User Friendliness
1. Abillity to create Anki decks instead of Csvs

### Desired features
1. Deconjugation to link long conjugation chains to the base verb and give more context
	a. Needs to link from verb to verb and questioning desirabillity
2. Working Cloze cards
3. Translation

### Known bugs
1. Sentences with 0 words will create files (should probably not exist)
2. I believe I know how to program

## Why Use This?
### Dogma
 Before I can explain why I believe this tool is valuable, I must explain my beliefs about the language learning process. I believe that due to the low inherit reward of the process of learning a language, we must find a highly motivating source. Whether it is love, cultural appreciation, a favorite book, or the allure of the end goal, whatever motivates us must be enough to endure a long and arduous process that delays gratification. I also believe that we learn languages best from the language itself out in its natural habitat. Music, movies, books, comics, animations, social media, articles etc. all the different sources that contain the very language we wish to speak. For natives of the Latin alphabet, reading another language with the latin alphabet is rather easy, often a beginner can get most of the sounds correct and remembering words isn't that much of an issue. With Kanji, hiragana, and katakana, the aspiring Japanese learner has a lot to hold in working memory for each sentence. Also, Kanji having many meanings and pronunciations means that evenatually this learner will have to expand upon their previous notes with new information. Beyond the belief of motivation and the belief that regular language is the best source to learn from, I'd rather not have much more dogma about language learning.  
 
  
So I wanted something that:

	1. Allows the learner to be motivated by their biggest motivating factor
	2. Facillitates and encourages engagement with the language
	3. Lowers the time it takes to gather data that one needs on any given kanji, word or sentence
	4. Allows for expansion with the knowledge of the language learner.

 
 A solution for 2 is a solution for 1, this is why the input is .txt files of user selected Japanese content.
 
 Automation of gathering data ~~using APIs~~ locally sourced data to not burden Jisho's servers and creating Flashcards with csvs is the solution for 3.
 
 Markdown is the answer for 4.

 
 And thus Japanese Content 2 Md And Anki was born.


## Usage tips

### 1. Mind Maps:

- As the Markdown was made with obsidian in mind, you can easily use its graph view to generate a mind map based on the inner links of all the notes files. This shows you all the connections for any given kanji, word, sentence, or piece of content.
### 2. Customize Translations and Definitions:

- As this project is meant to be open source, I don't have much for translation options, but I want to add translation capabillities. Currently this version slings every definition at every word, I suggest personalizing each note that you come across with links and your own ttranslations.


### 3. Use pictures for objects:

- Unless you need to practice typing because a picture is worth a thousand words. These are markdown files, they're meant for you to put images or anything else you can get working in them. Many Markdown previewers will show any images you add to a markdown file via links. Use this to your advantage.

### 4. Use content you love:

- As I mentioned motivation before, content that you love is going to be a much better use of your learning hours than something you don't.
- My suggestion for learning from sources you don't love is to study desirable content first and see what from the undesirable content you already know. If you already know most of it, it might be highly motivating to go through the process of learning that little extra bit!

## How to Study with Japanese Content 2 Md And Anki

After running the script and generating your Japanese language learning notes, you can optimize your study process using the generated content. Here's a suggested study approach:

### 1. Review Content Markdown Files:
- These are lists for finding something specific
	- **Content.md**
	- **Kanji.md**
	- **Words.md:**
	- **Sentences.md:**
- These are effectively the database of your knowledge 

- I suggest going through content from Content.md and scrolling down to the Sentence links and opening the first one you don't understand
- From there I suggest studying in this manner, assuming you have no knowledge:
	- Sentence => ...Kanji => Word=> Next Kanji => Word => ... => Sentence
	- You may or may not have to adjust the words as the parser is usually right but sometimes it's wrong
	- ~~You may also have to adjust the links as sometimes a word will link to itself instead of its kanji~~
		- this was a bug and it has been resolved
	- The translation of the sentence, or currently the lack thereof, might not be suitable for the Content or for your mind, use translation services [DeepL](https://www.deepl.com/translator), [Jisho](https://jisho.org/), [Google Translate](https://translate.google.com/), (it's only bad in isolation), [ChatGPT](https://chat.openai.com/)
	- Continue doing this until the end of your study session or until you've done every sentence.

### 2. Create Anki Flashcards (Optional):
- You 'know' and remember the content that you've studied. Now's the time you'd want to make flashcards especially since you've corrected the sentences.
- If you haven't corrected the markdown notes, now's the time also now's the time to customize the markdown notes as they dictate what goes on the cards
- Currently, the 3rd line is the start of the front card as the format was meant to be kept in line with [Obsidian to Anki](https://github.com/Pseudonium/Obsidian_to_Anki), becareful
- If you're using Anki for flashcards, import the generated `.csv` files from the `Notes\Japanese Notes\CSV` directory
- Use the flashcards to reinforce your memory

### 3. Consistent Practice:

- Either use the flashcards or the content itself to remember everything you've learned before

## File Structure
### Directories
 - Notes\Japanese Notes\Content: Directory for individual content markdown files.
 - Notes\Japanese Notes\Kanji: Directory for individual kanji markdown files.
 - Notes\Japanese Notes\Sentences: Directory for individual sentences markdown files.
 - Notes\Japanese Notes\Words: Directory for individual words markdown files.
 - Notes\Japanese Notes\CSV: Directory for generated CSV files.
### Markdown Files
 - Notes\Japanese Notes\Content.md: Main content markdown file.
 - Notes\Japanese Notes\Kanji.md: Kanji notes markdown file.
 - Notes\Japanese Notes\Sentences.md: Sentences notes markdown file.
 - Notes\Japanese Notes\Words.md: Words notes markdown file.


## How to Contribute
- All pull requests are welcome :)
- please create an issue if you find a problem or desire a feature
