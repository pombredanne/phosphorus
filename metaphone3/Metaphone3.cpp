/**
 *
 * Metaphone 3
 * VERSION 2.1.3
 * by Lawrence Philips
 *
 * Metaphone 3 is designed to return an *approximate* phonetic key (and an alternate
 * approximate phonetic key when appropriate) that should be the same for english
 * words, and most names familiar in the United States, that are pronounced *similarly*.
 * The key value is *not* intended to be an *exact* phonetic, or even phonemic,
 * representation of the word. This is because a certain degree of 'fuzziness' has
 * proven to be useful in compensating for variations in pronounciation, as well as
 * mis-heard pronounciations. For example, although americans are not usually aware of it,
 * the letter 's' is normally pronounced 'z' at the end of words such as "sounds".
 *
 * The 'approximate' aspect of the encoding is implemented according to the following rules:
 *
 * (1) All vowels are encoded to the same value - 'A'. If the parameter encodeVowels
 * is set to false, only *initial* vowels will be encoded at all. If encodeVowels is set
 * to true, 'A' will be encoded at all places in the word that any vowels are normally
 * pronounced. 'W' as well as 'Y' are treated as vowels. Although there are differences in
 * the pronounciation of 'W' and 'Y' in different circumstances that lead to their being
 * classified as vowels under some circumstances and as consonants in others, for the puposes
 * of the 'fuzzyness' component of the Soundex and Metaphone family of algorithms they will
 * be always be treated here as vowels.
 *
 * (2) Voiced and un-voiced consonant pairs are mapped to the same encoded value. This
 * means that:
 * 'D' and 'T' -> 'T'
 * 'B' and 'P' -> 'P'
 * 'G' and 'K' -> 'K'
 * 'Z' and 'S' -> 'S'
 * 'V' and 'F' -> 'F'
 *
 * - In addition to the above voiced/unvoiced rules, 'CH' and 'SH' -> 'X', where 'X'
 * represents the "-SH-" and "-CH-" sounds in Metaphone 3 encoding.
 *
 * - Also, the sound that is spelled as "TH" in english is encoded to '0' (zero symbol). (Although
 * americans are not usually aware of it, "TH" is pronounced in a voiced (e.g. "that") as
 * well as an unvoiced (e.g. "theater") form, which are natually mapped to the same encoding.)
 *
 * The encodings in this version of Metaphone 3 are according to pronounciations common in the
 * United States. This means that they will be innacurate for consonant pronounciations that
 * are different in the United Kingdom, for example "tube" -> "CHOOBE" -> XAP rather than american TAP.
 *
 * Metaphone 3 was preceded by by Soundex, patented in 1919, and Metaphone and Double Metaphone,
 * developed by Lawrence Philips. All of these algorithms resulted in a significant number of
 * incorrect encodings. Metaphone3 was tested against a database of about 100 thousand english words,
 * names common in the United States, and non-english words found in publications in the United States,
 * with an emphasis on words that are commonly mispronounced, prepared by the Moby Words website,
 * but with the Moby Words 'phonetic' encodings algorithmically mapped to Double Metaphone encodings.
 * Metaphone3 increases the accuracy of encoding of english words, common names, and non-english
 * words found in american publications from the 89% for Double Metaphone, to over 98%.
 *
 * DISCLAIMER:
 * Anthropomorphic Software LLC claims only that Metaphone 3 will return correct encodings,
 * within the 'fuzzy' definition of correct as above, for a very high percentage of correctly
 * spelled english and commonly recognized non-english words. Anthropomorphic Software LLC
 * warns the user that a number of words remain incorrectly encoded, that misspellings may not
 * be encoded 'properly', and that people often have differing ideas about the pronounciation
 * of a word. Therefore, Metaphone 3 is not guaranteed to return correct results everytime, and
 * so a desired target word may very well be missed. Creators of commercial products should
 * keep in mind that systems like Metaphone 3 produce a 'best guess' result, and should
 * condition the expectations of end users accordingly.
 *
 * NOTE FOR PROGRAMMERS intending to make changes to this source code:
 * There are a number of places where the code indexes into the input string character array
 * without explicit bounds checking, in this way: m_inWord[m_current + 1]. If you examine the code
 * carefully, you will be able to see that currently this is always done safely - subtracting from
 * m_current is always bounds checked, m_current + 1 will at worst access the terminating null, and
 * all cases of m_current + 2 occur in places where it has already been determined that m_current + 1
 * accesses a valid alpha character which is before the terminating null. However, any changes to the
 * code run the risk of exposing this array accessing to errors - please be careful and verify that
 * you are not running the risk of invalid memory accesses!
 */

// #include "stdafx.h"
#include "Metaphone3.h"
#include <string.h>
#include <stdarg.h>

#define AND &&
#define OR ||

////////////////////////////////////////////////////////////////////////////////
// Metaphone3 class definition
////////////////////////////////////////////////////////////////////////////////

/**
 * Constructor, default. This constructor is most convenient when
 * encoding more than one word at a time. New words to encode can
 * be set using SetWord(char *).
 *
 */
Metaphone3::Metaphone3()
{
	Init();
}

/**
 * Constructor, parameterized. The Metaphone3 object will
 * be initialized with the incoming string, and can be called
 * on to encode this string. This constructor is most convenient
 * when only one word needs to be encoded.
 *
 * @param in pointer to char string of word to be encoded.
 *
 */
Metaphone3::Metaphone3(const char* in)
{
	Init();

	m_inWord = in;
	std::transform(m_inWord.begin(), m_inWord.end(),
			m_inWord.begin(), (int(*)(int)) std::toupper);
	ConvertExtendedASCIIChars();
	m_inWord += ' ';
}


/**
 * Initialize class variables
 *
 */
void Metaphone3::Init()
{
	m_inWord.clear();
	m_primary.clear();
    m_secondary.clear();

    m_metaphLength = DEFAULT_MAX_KEY_LENGTH;
    m_encodeVowels = false;
	m_encodeExact = false;

}

/**
 * Sets word to be encoded.
 *
 * @param in pointer to EXTERNALLY ALLOCATED char string of
 * the word to be encoded.
 *
 */
void Metaphone3::SetWord(const char* in)
{
	m_inWord.clear();
	m_inWord = in;
	std::transform(m_inWord.begin(), m_inWord.end(),
			m_inWord.begin(), (int(*)(int)) std::toupper);
	ConvertExtendedASCIIChars();
	m_inWord += ' ';
}

/**
 * Maps extended ASCII characters in our input
 * string to uppercase
 *
 */
void Metaphone3::ConvertExtendedASCIIChars()
{
    m_length = m_inWord.length();
	for(int i = 0; i < m_length; i++)
	{
        char c = m_inWord[i];
		if(c >= 'à' && c <= 'þ')
		{
			m_inWord[i] = ('À' + c - 'à');
		}
		if(c >= 'š' && c <= 'ž')
		{
			m_inWord[i] = ('Ž' + c - 'š');
		}
		else if(c == 'ÿ')
		{
			m_inWord[i] = 'Ÿ';
		}
	}
}

/**
 * Sets length allocated for output keys.
 * If incoming number is greater than maximum allowable
 * length returned by GetMaximumKeyLength(), set key length
 * to maximum key length and return false;  otherwise, set key
 * length to parameter value and return true.
 *
 * @param inKeyLength new length of key.
 * @return true if able to set key length to requested value.
 *
*/
bool Metaphone3::SetKeyLength(unsigned short inKeyLength)
{
    if(inKeyLength < 1)
	{
		// can't have that -
		// no room for terminating null
		inKeyLength = 1;
	}

	if(inKeyLength > MAX_KEY_ALLOCATION)
    {
        m_metaphLength = MAX_KEY_ALLOCATION;
        return false;
    }

     m_metaphLength = inKeyLength;
     return true;
}

/**
 * Adds an encoding character to the encoded key value string - one parameter version.
 *
 * @param main primary encoding character to be added to encoded key string.
 *
 */
inline void Metaphone3::MetaphAdd(const char* main)
{
    if(!((*main == 'A')
    	AND (m_primary.length() > 0)
    	AND (m_primary[m_primary.length() - 1] == 'A')))
    {
		m_primary.append(main);
	}

    if(!((*main == 'A')
    	AND (m_secondary.length() > 0)
    	AND (m_secondary[m_secondary.length() - 1] == 'A')))
    {
		m_secondary.append(main);
	}
}

/**
 * Adds an encoding character to the encoded key value string - two parameter version
 *
 * @param main primary encoding character to be added to encoded key string
 * @param alt alternative encoding character to be added to encoded alternative key string
 *
 */
inline void Metaphone3::MetaphAdd(const char* main, const char* alt)
{
    if(!((*main == 'A')
    	AND (m_primary.length() > 0)
    	AND (m_primary[m_primary.length() - 1] == 'A')))
    {
 		m_primary.append(main);
	}

    if(!((*alt == 'A')
    	AND (m_secondary.length() > 0)
    	AND (m_secondary[m_secondary.length() - 1] == 'A')))
    {
		if(*alt)
		{
			m_secondary.append(alt);
		}
	}
}

/**
 * Adds an encoding character to the encoded key value string - Exact/Approx version
 *
 * @param mainExact primary encoding character to be added to encoded key string if
 * m_encodeExact is set
 *
 * @param altExact alternative encoding character to be added to encoded alternative
 * key string if m_encodeExact is set
 *
 * @param main primary encoding character to be added to encoded key string
 *
 * @param alt alternative encoding character to be added to encoded alternative key string
 *
 */
inline void Metaphone3::MetaphAddExactApprox(const char* mainExact, const char* altExact, const char* main, const char* alt)
{
	if(m_encodeExact)
	{
		MetaphAdd(mainExact, altExact);
	}
	else
	{
		MetaphAdd(main, alt);
	}
}

/**
 * Adds an encoding character to the encoded key value string - Exact/Approx version
 *
 * @param mainExact primary encoding character to be added to encoded key string if
 * m_encodeExact is set
 *
 * @param main primary encoding character to be added to encoded key string
 *
 */
inline void Metaphone3::MetaphAddExactApprox(const char* mainExact, const char* main)
{
	if(m_encodeExact)
	{
		MetaphAdd(mainExact);
	}
	else
	{
		MetaphAdd(main);
	}
}
/**
 * Test for close front vowels
 *
 * @return true if close front vowel
 */
bool Metaphone3::Front_Vowel(int at)
{
	if((m_inWord[at] == 'E') OR (m_inWord[at] == 'I') OR (m_inWord[at] == 'Y'))
	{
		return true;
	}

	return false;
}

/**
 * Detect names or words that begin with spellings
 * typical of german or slavic words, for the purpose
 * of choosing alternate pronunciations correctly
 *
 */
bool Metaphone3::SlavoGermanic()
{
	if(StringAt(0, 3, "SCH", "")
		OR StringAt(0, 2, "SW", "")
		OR (m_inWord[0] == 'J')
		OR (m_inWord[0] == 'W'))
	{
		return true;
	}

	return false;
}

/**
 * Tests if character is a vowel
 *
 * @param inChar character to be tested in string to be encoded
 * @return true if character is a vowel, false if not
 *
 */
bool Metaphone3::IsVowel(char inChar)
{
    if((inChar == 'A')
		OR (inChar == 'E')
		OR (inChar == 'I')
		OR (inChar == 'O')
		OR (inChar == 'U')
		OR (inChar == 'Y')
		OR (inChar == 'À')
		OR (inChar == 'Á')
		OR (inChar == 'Â')
		OR (inChar == 'Ã')
		OR (inChar == 'Ä')
		OR (inChar == 'Å')
		OR (inChar == 'Æ')
		OR (inChar == 'È')
		OR (inChar == 'É')
 		OR (inChar == 'Ê')
		OR (inChar == 'Ë')
		OR (inChar == 'Ì')
		OR (inChar == 'Í')
		OR (inChar == 'Î')
 		OR (inChar == 'Ï')
		OR (inChar == 'Ò')
		OR (inChar == 'Ó')
		OR (inChar == 'Ô')
		OR (inChar == 'Õ')
		OR (inChar == 'Ö')
		OR (inChar == 'Œ')
		OR (inChar == 'Ø')
		OR (inChar == 'Ù')
		OR (inChar == 'Ú')
		OR (inChar == 'Û')
		OR (inChar == 'Ü')
		OR (inChar == 'Ý')
		OR (inChar == 'Ÿ'))
	{
        return true;
	}

    return false;
}

/**
 * Tests if character in the input string is a vowel
 *
 * @param at position of character to be tested in string to be encoded
 * @return true if character is a vowel, false if not
 *
 */
bool Metaphone3::IsVowel(int at)
{
    if((at < 0) OR (at >= m_length))
	{
        return false;
	}

    char it = m_inWord[at];

    if(IsVowel(it))
	{
        return true;
	}

    return false;
}

/**
 * Skips over vowels in a string. Has exceptions for skipping consonants that
 * will not be encoded.
 *
 * @param at position, in string to be encoded, of character to start skipping from
 *
 * @return position of next consonant in string to be encoded
 */
int Metaphone3::SkipVowels(int at)
{
    if(at < 0)
	{
        return 0;
	}

    if(at >= m_length)
	{
        return m_length;
	}

    char it = m_inWord[at];

    while(IsVowel(it)
			OR (it == 'W'))
    {
        if(StringAt(at, 4, "WICZ", "WITZ", "WIAK", "")
			OR StringAt((at - 1), 5, "EWSKI", "EWSKY", "OWSKI", "OWSKY", "")
			OR (StringAt(at, 5, "WICKI", "WACKI", "") && ((at + 4) == m_last)))
        {
            break;
        }

        at++;
        if(((m_inWord[at - 1] == 'W') AND (m_inWord[at] == 'H'))
            AND !(StringAt(at, 3, "HOP", "")
                  OR StringAt(at, 4, "HIDE", "HARD", "HEAD", "HAWK", "HERD", "HOOK", "HAND", "HOLE", "")
                  OR StringAt(at, 5, "HEART", "HOUSE", "HOUND", "")
                  OR StringAt(at, 6, "HAMMER", "")))
		{
            at++;
		}
        it = m_inWord[at];
    }

    return at;
}

/**
 * Advanced counter m_current so that it indexes the next character to be encoded
 *
 * @param ifNotEncodeVowels number of characters to advance if not encoding internal vowels
 * @param ifEncodeVowels number of characters to advance if encoding internal vowels
 *
 */
void Metaphone3::AdvanceCounter(int ifNotEncodeVowels, int ifEncodeVowels)
{
	if(!m_encodeVowels)
	{
		m_current += ifNotEncodeVowels;
	}
	else
	{
		m_current += ifEncodeVowels;
	}
}

/**
 * Tests equality of substring in word to be encoded, starting at position 'start',
 * and having a length of 'length', with a variable number of character string
 * parameters, TERMINATED BY A NULL CHARACTER
 *
 * @return true if any of the variable number of character string parameters match
 * exactly the substring at the designated position in the string to be encoded
 *
 */
bool Metaphone3::StringAt(int start, int length, ... )
{
    char    *test;
    int     t = 0, w, end;
    bool    match = true;

	// check substring bounds
    if((start < 0)
		OR (start > (m_length - 1))
		OR ((start + length - 1) > (m_length - 1)))
	{
        return false;
	}

    va_list sstrings;
    va_start(sstrings, length);

    test = va_arg(sstrings, char*);

    while(test[0] != '\0')
    {
        match = true;
        w = start; t = 0;
        end = start + length;

        while(w < end)
        {
            if(m_inWord[w] != test[t])
            {
                match = false;
                break;
            }
            w++; t++;
        }

        if(match)
        {
            return true;
        }

        test = va_arg(sstrings, char*);
    }

    va_end(sstrings);

    return false;
}

/**
 * Tests whether the word is the root or a regular english inflection
 * of it, e.g. "ache", "achy", "aches", "ached", "aching", "achingly"
 * This is for cases where we want to match only the root and corresponding
 * inflected forms, and not completely different words which may have the
 * same substring in them.
 */
bool Metaphone3::RootOrInflections(string inWord, string root)
{
	int len = root.length();
	string test;

	test = root + "S";
	if((inWord == root)
		OR (inWord == test))
	{
		return true;
	}

	if(root[len - 1] != 'E')
	{
		test = root + "ES";
	}

	if(inWord == test)
	{
		return true;
	}

	if(root[len - 1] != 'E')
	{
		test = root + "ED";
	}
	else
	{
		test = root + "D";
	}

	if(inWord == test)
	{
		return true;
	}

	if(root[len - 1] == 'E')
	{
		root.resize(len - 1);
	}

	test = root + "ING";
	if(inWord == test)
	{
		return true;
	}

	test = root + "INGLY";
	if(inWord == test)
	{
		return true;
	}

	test = root + "Y";
	if(inWord == test)
	{
		return true;
	}

	return false;
}

/**
 * Encodes input string to one or two key values according to Metaphone 3 rules.
 *
 */
void Metaphone3::Encode()
{
    flag_AL_inversion = false;

    m_current = 0;

	m_primary.clear();
    m_secondary.clear();

    if(m_length < 1)
	{
        return;
	}

    //zero based index
	m_last = m_length - 1;

    ///////////main loop//////////////////////////
	while(!(m_primary.length() > m_metaphLength) AND !(m_secondary.length() > m_metaphLength))
    {
        if(m_current >= m_length)
		{
            break;
		}

        switch(m_inWord[m_current])
        {
            case 'B':

				Encode_B();
                break;

            case 'ß':
			case 'Ç':

                MetaphAdd("S");
                m_current++;
                break;

            case 'C':

				Encode_C();
                break;

            case 'D':

				Encode_D();
                break;

            case 'F':

				Encode_F();
                break;

            case 'G':

				Encode_G();
                break;

            case 'H':

				Encode_H();
                break;

            case 'J':

				Encode_J();
                break;

            case 'K':

				Encode_K();
                break;

            case 'L':

				Encode_L();
                break;

            case 'M':

				Encode_M();
                break;

            case 'N':

				Encode_N();
                break;

            case 'Ñ':

                MetaphAdd("N");
                m_current++;
                break;

            case 'P':

				Encode_P();
                break;

            case 'Q':

				Encode_Q();
                break;

            case 'R':

				Encode_R();
				break;

            case 'S':

				Encode_S();
                break;

            case 'T':

				Encode_T();
                break;

            case 'Ð': // eth
			case 'Þ': // thorn

                MetaphAdd("0");
                m_current++;
                break;

           case 'V':

				Encode_V();
                break;

            case 'W':

				Encode_W();
                break;

            case 'X':

				Encode_X();
                break;

            case 'Š':

                MetaphAdd("X");
                m_current++;
                break;

			case 'Ž':

                MetaphAdd("S");
                m_current++;
                break;

            case 'Z':

                Encode_Z();
                break;

            default:

				if(IsVowel(m_inWord[m_current]))
				{
					Encode_Vowels();
					break;
				}

                m_current++;
        }
    }

    //only give back m_metaphLength number of chars in m_metaph
	if(m_primary.length() > m_metaphLength)
    {
		m_primary[m_metaphLength]  = '\0';
    }

	if(m_secondary.length() > m_metaphLength)
    {
		m_secondary[m_metaphLength]  = '\0';
    }

	// it is possible for the two metaphs to be the same
	// after truncation. lose the second one if so
	if(m_primary == m_secondary)
	{
		m_secondary.clear();
	}
}

/**
 * Encodes all initial vowels to A.
 *
 * Encodes non-initial vowels to A if m_encodeVowels is true
 *
 *
*/
void Metaphone3::Encode_Vowels()
{
	if(m_current == 0)
	{
		// all init vowels map to 'A'
		// as of Double Metaphone
		MetaphAdd("A");
	}
	else if(m_encodeVowels)
	{
		if(m_inWord[m_current] != 'E')
		{
			if(Skip_Silent_UE())
			{
				return;
			}

			if (O_Silent())
			{
				m_current++;
				return;
			}

			// encode all vowels and
			// diphthongs to the same value
			MetaphAdd("A");
		}
		else
		{
			Encode_E_Pronounced();
		}
	}

	if(!(!IsVowel(m_current - 2) AND StringAt((m_current - 1), 4, "LEWA", "LEWO", "LEWI", "")))
	{
		m_current = SkipVowels(m_current);
	}
	else
	{
		m_current++;
	}
}

/**
 * Encodes cases where non-initial 'e' is pronounced, taking
 * care to detect unusual cases from the greek.
 *
 * Only executed if non initial vowel encoding is turned on
 *
 *
 */
void Metaphone3::Encode_E_Pronounced()
{
	// special cases with two pronounciations
	// 'agape' 'lame' 'resume'
	if((StringAt(0, 4, "LAME", "SAKE", "PATE", "") && (m_length == 4))
		OR (StringAt(0, 5, "AGAPE", "") && (m_length == 5))
		OR ((m_current == 5) AND StringAt(0, 6, "RESUME", "")))
	{
		MetaphAdd("", "A");
		return;
	}

	// special case "inge" => 'INGA', 'INJ'
	if(StringAt(0, 4, "INGE", "")
		AND (m_length == 4))
	{
		MetaphAdd("A", "");
		return;
	}

	// special cases with two pronunciations
	// special handling due to the difference in
	// the pronunciation of the '-D'
	if((m_current == 5) AND StringAt(0, 7, "BLESSED", "LEARNED", ""))
	{
		MetaphAddExactApprox("D", "AD", "T", "AT");
		m_current += 2;
		return;
	}

	// encode all vowels and diphthongs to the same value
	if((!E_Silent()
			AND !flag_AL_inversion
			AND !Silent_Internal_E())
		OR E_Pronounced_Exceptions())
	{
		MetaphAdd("A");
	}

	// now that we've visited the vowel in question
	flag_AL_inversion = false;
}

/**
 * Tests for cases where non-initial 'o' is not pronounced
 * Only executed if non initial vowel encoding is turned on
 *
 * @return true if encoded as silent - no addition to m_metaph key
 *
 */
bool Metaphone3::O_Silent()
{
	// if "iron" at beginning or end of word and not "irony"
	if ((m_inWord[m_current] == 'O')
		&& StringAt((m_current - 2), 4, "IRON", ""))
	{
		if ((StringAt(0, 4, "IRON", "")
			|| (StringAt((m_current - 2), 4, "IRON", "")
				&& (m_last == (m_current + 1))))
			&& !StringAt((m_current - 2), 6, "IRONIC", ""))
		{
			return true;
		}
	}

	return false;
}

/**
 * Tests and encodes cases where non-initial 'e' is never pronounced
 * Only executed if non initial vowel encoding is turned on
 *
 * @return true if encoded as silent - no addition to m_metaph key
 *
*/
bool Metaphone3::E_Silent()
{
	if(E_Pronounced_At_End())
	{
		return false;
	}

	// 'e' silent when last letter, altho
	if((m_current == m_last)
		// also silent if before plural 's'
		// or past tense or participle 'd', e.g.
		// 'grapes' and 'banished' => PNXT
		OR ((StringAt(m_last, 1, "S", "D", "")
		AND (m_current > 1)
		AND ((m_current + 1) == m_last)
			// and not e.g. "nested", "rises", or "pieces" => RASAS
			AND !(StringAt((m_current - 1), 3, "TED", "SES", "CES", "")
				  OR StringAt(0, 9, "ANTIPODES", "ANOPHELES", "")
				  OR StringAt(0, 8, "MOHAMMED", "MUHAMMED", "MOUHAMED", "")
				  OR StringAt(0, 7, "MOHAMED", "")
				  OR StringAt(0, 6, "NORRED", "MEDVED", "MERCED", "ALLRED", "KHALED", "RASHED", "MASJED", "")
				  OR StringAt(0, 5, "JARED", "AHMED", "HAMED", "JAVED", "")
				  OR StringAt(0, 4, "ABED", "IMED", ""))))
			// e.g.  'wholeness', 'boneless', 'barely'
			OR (StringAt((m_current + 1), 4, "NESS", "LESS", "") && ((m_current + 4) == m_last))
			OR (StringAt((m_current + 1), 2, "LY", "") && ((m_current + 2) == m_last)
					AND !StringAt(0, 6, "CICELY", "")))
	{
		return true;
	}

	return false;
}

/**
 * Tests for words where an 'E' at the end of the word
 * is pronounced
 *
 * special cases, mostly from the greek, spanish, japanese,
 * italian, and french words normally having an acute accent.
 * also, pronouns and articles
 *
 * Many Thanks to ali, QuentinCompson, JeffCO, ToonScribe, Xan,
 * Trafalz, and VictorLaszlo, all of them atriots from the Eschaton,
 * for all their fine contributions!
 *
 * @return true if 'E' at end is pronounced
 *
*/
bool Metaphone3::E_Pronounced_At_End()
{
	if((m_current == m_last)
		AND (StringAt((m_current - 6), 7, "STROPHE", "")
		// if a vowel is before the 'E', vowel eater will have eaten it.
		//otherwise, consonant + 'E' will need 'E' pronounced
		OR (m_length == 2)
		OR ((m_length == 3) AND !IsVowel(0))
		// these german name endings can be relied on to have the 'e' pronounced
		OR (StringAt((m_last - 2), 3, "BKE", "DKE", "FKE", "KKE", "LKE",
									 "NKE", "MKE", "PKE", "TKE", "VKE", "ZKE", "")
			AND !StringAt(0, 5, "FINKE", "FUNKE", "")
			AND !StringAt(0, 6, "FRANKE", ""))
		OR StringAt((m_last - 4), 5, "SCHKE", "")
		OR (StringAt(0, 4, "ACME", "NIKE", "CAFE", "RENE", "LUPE", "JOSE", "ESME", "") AND (m_length == 4))
		OR (StringAt(0, 5, "LETHE", "CADRE", "TILDE", "SIGNE", "POSSE", "LATTE", "ANIME", "DOLCE", "CROCE",
							"ADOBE", "OUTRE", "JESSE", "JAIME", "JAFFE", "BENGE", "RUNGE",
							"CHILE", "DESME", "CONDE", "URIBE", "LIBRE", "ANDRE", "") AND (m_length == 5))
		OR (StringAt(0, 6, "HECATE", "PSYCHE", "DAPHNE", "PENSKE", "CLICHE", "RECIPE",
						   "TAMALE", "SESAME", "SIMILE", "FINALE", "KARATE", "RENATE", "SHANTE",
						   "OBERLE", "COYOTE", "KRESGE", "STONGE", "STANGE", "SWAYZE", "FUENTE",
						   "SALOME", "URRIBE", "") AND (m_length == 6))
		OR (StringAt(0, 7, "ECHIDNE", "ARIADNE", "MEINEKE", "PORSCHE", "ANEMONE", "EPITOME",
							"SYNCOPE", "SOUFFLE", "ATTACHE", "MACHETE", "KARAOKE", "BUKKAKE",
							"VICENTE", "ELLERBE", "VERSACE", "") AND (m_length == 7))
		OR (StringAt(0, 8, "PENELOPE", "CALLIOPE", "CHIPOTLE", "ANTIGONE", "KAMIKAZE", "EURIDICE",
						   "YOSEMITE", "FERRANTE", "") AND (m_length == 8))
		OR (StringAt(0, 9, "HYPERBOLE", "GUACAMOLE", "XANTHIPPE", "") AND (m_length == 9))
		OR (StringAt(0, 10, "SYNECDOCHE", "") AND (m_length == 10))))
	{
		return true;
	}

	return false;
}

/**
 * Detect internal silent 'E's e.g. "roseman",
 * "firestone"
 *
 */
bool Metaphone3::Silent_Internal_E()
{
	// 'olesen' but not 'olen'
	if((StringAt(0, 3, "OLE", "")
				AND E_Silent_Suffix(3) AND !E_Pronouncing_Suffix(3))
	   OR (StringAt(0, 4, "BARE", "FIRE", "FORE", "GATE", "HAGE", "HAVE",
			             "HAZE", "HOLE", "CAPE", "HUSE", "LACE", "LINE",
			             "LIVE", "LOVE", "MORE", "MOSE", "MORE", "NICE",
			             "RAKE", "ROBE", "ROSE", "SISE", "SIZE", "WARE",
			             "WAKE", "WISE", "WINE", "")
				AND E_Silent_Suffix(4) AND !E_Pronouncing_Suffix(4))
	   OR (StringAt(0, 5, "BLAKE", "BRAKE", "BRINE", "CARLE", "CLEVE", "DUNNE",
			   			 "HEDGE", "HOUSE", "JEFFE", "LUNCE", "STOKE", "STONE",
			   			 "THORE", "WEDGE", "WHITE", "")
				 AND E_Silent_Suffix(5) AND !E_Pronouncing_Suffix(5))
	   OR (StringAt(0, 6, "BRIDGE", "CHEESE", "")
				 AND E_Silent_Suffix(6) AND !E_Pronouncing_Suffix(6))
	   OR StringAt(0, 7, "CHARLES", ""))
	{
		return true;
	}

	return false;
}

/**
 * Detect conditions required
 * for the 'E' not to be pronounced
 *
 */
bool Metaphone3::E_Silent_Suffix(int at)
{
	if((m_current == (at - 1))
			AND (m_length > (at + 1))
			AND (IsVowel((at + 1))
			OR (StringAt(at, 2, "ST", "SL", "")
				AND (m_length > (at + 2)))))
	{
		return true;
	}

	return false;
}

/**
 * Detect endings that will
 * cause the 'e' to be pronounced
 *
 */
bool Metaphone3::E_Pronouncing_Suffix(int at)
{
	// e.g. 'bridgewood' - the other vowels will get eaten
	// up so we need to put one in here
	if((m_length == (at + 4)) AND StringAt(at, 4, "WOOD", ""))
	{
		return true;
	}

	// same as above
	if((m_length == (at + 5)) && StringAt(at, 5, "WATER", "WORTH", ""))
	{
		return true;
	}

	// e.g. 'bridgette'
	if((m_length == (at + 3)) AND StringAt(at, 3, "TTE", "LIA", "NOW", "ROS", "RAS", ""))
	{
		return true;
	}

	// e.g. 'olena'
	if((m_length == (at + 2)) AND StringAt(at, 2, "TA", "TT", "NA", "NO", "NE",
												  "RS", "RE", "LA", "AU", "RO", "RA", ""))
	{
		return true;
	}

	// e.g. 'bridget'
	if((m_length == (at + 1)) AND StringAt(at, 1, "T", "R", ""))
	{
		return true;
	}

	return false;
}

/**
 * Exceptions where 'E' is pronounced where it
 * usually wouldn't be, and also some cases
 * where 'LE' transposition rules don't apply
 * and the vowel needs to be encoded here
 *
 * @return true if 'E' pronounced
 *
 */
bool Metaphone3::E_Pronounced_Exceptions()
{
	// greek names e.g. "herakles" or hispanic names e.g. "robles", where 'e' is pronounced, other exceptions
	if((((m_current + 1) == m_last)
		AND (StringAt((m_current - 3), 5, "OCLES", "ACLES", "AKLES", "")
			OR StringAt(0, 4, "INES", "")
			OR StringAt(0, 5, "LOPES", "ESTES", "GOMES", "NUNES", "ALVES", "ICKES",
							  "INNES", "PERES", "WAGES", "NEVES", "BENES", "DONES", "")
			OR StringAt(0, 6, "CORTES", "CHAVES", "VALDES", "ROBLES", "TORRES", "FLORES", "BORGES",
							  "NIEVES", "MONTES", "SOARES", "VALLES", "GEDDES", "ANDRES", "VIAJES",
							  "CALLES", "FONTES", "HERMES", "ACEVES", "BATRES", "MATHES", "")
			OR StringAt(0, 7, "DELORES", "MORALES", "DOLORES", "ANGELES", "ROSALES", "MIRELES", "LINARES",
							  "PERALES", "PAREDES", "BRIONES", "SANCHES", "CAZARES", "REVELES", "ESTEVES",
							  "ALVARES", "MATTHES", "SOLARES", "CASARES", "CACERES", "STURGES", "RAMIRES",
							  "FUNCHES", "BENITES", "FUENTES", "PUENTES", "TABARES", "HENTGES", "VALORES", "")
			OR StringAt(0, 8, "GONZALES", "MERCEDES", "FAGUNDES", "JOHANNES", "GONSALES", "BERMUDES",
							  "CESPEDES", "BETANCES", "TERRONES", "DIOGENES", "CORRALES", "CABRALES",
							  "MARTINES", "GRAJALES", "")
			OR StringAt(0, 9, "CERVANTES", "FERNANDES", "GONCALVES", "BENEVIDES", "CIFUENTES", "SIFUENTES",
							  "SERVANTES", "HERNANDES", "BENAVIDES", "")
			OR StringAt(0, 10, "ARCHIMEDES", "CARRIZALES", "MAGALLANES", "")))
		OR StringAt(m_current - 2, 4, "FRED", "DGES", "DRED", "GNES", "")
		OR StringAt((m_current - 5), 7, "PROBLEM", "RESPLEN", "")
		OR StringAt((m_current - 4), 6, "REPLEN", "")
		OR StringAt((m_current - 3), 4, "SPLE", ""))
	{
		return true;
	}

	return false;
}

/**
 * Encodes "-UE".
 *
 * @return true if encoding handled in this routine, false if not
 */
bool Metaphone3::Skip_Silent_UE()
{
	// always silent except for cases listed below
	if((StringAt((m_current - 1), 3, "QUE", "GUE", "")
		AND !StringAt(0, 8, "BARBEQUE", "PALENQUE", "APPLIQUE", "")
		// '-que' cases ususally french but missing the acute accent
		AND !StringAt(0, 6, "RISQUE", "")
		AND !StringAt((m_current - 3), 5, "ARGUE", "SEGUE", "")
		AND !StringAt(0, 7, "PIROGUE", "ENRIQUE", "")
		AND !StringAt(0, 10, "COMMUNIQUE", ""))
		AND (m_current > 1)
			AND (((m_current + 1) == m_last)
				OR StringAt(0, 7, "JACQUES", "")))
	{
		m_current = SkipVowels(m_current);
		return true;
	}

	return false;
}

/**
 * Encodes 'B'
 *
 *
 */
void Metaphone3::Encode_B()
{
	if(Encode_Silent_B())
	{
		return;
	}

	// "-mb", e.g", "dumb", already skipped over under
	// 'M', altho it should really be handled here...
	MetaphAddExactApprox("B", "P");

	if((m_inWord[m_current + 1] == 'B')
		OR ((m_inWord[m_current + 1] == 'P')
		AND ((m_current + 1 < m_last) AND (m_inWord[m_current + 2] != 'H'))))
	{
		m_current += 2;
	}
	else
	{
		m_current++;
	}
}

/**
 * Encodes silent 'B' for cases not covered under "-mb-"
 *
 *
 * @return true if encoding handled in this routine, false if not
 *
*/
bool Metaphone3::Encode_Silent_B()
{
	//'debt', 'doubt', 'subtle'
	if(StringAt((m_current - 2), 4, "DEBT", "")
		OR StringAt((m_current - 2), 5, "SUBTL", "")
		OR StringAt((m_current - 2), 6, "SUBTIL", "")
		OR StringAt((m_current - 3), 5, "DOUBT", ""))
	{
		MetaphAdd("T");
		m_current += 2;
		return true;
	}

	return false;
}

/**
 * Encodes 'C'
 *
 */
void Metaphone3::Encode_C()
{

	if(Encode_Silent_C_At_Beginning()
		OR Encode_CA_To_S()
		OR Encode_CO_To_S()
		OR Encode_CH()
		OR Encode_CCIA()
		OR Encode_CC()
		OR Encode_CK_CG_CQ()
		OR Encode_C_Front_Vowel()
		OR Encode_Silent_C()
		OR Encode_CZ())
	{
		return;
	}

	// give an 'etymological' 2nd
	// encoding for "kovacs" so
	// that it matches "kovach"
	if(StringAt(0, 6, "KOVACS", ""))
	{
		MetaphAdd("KS", "X");
		m_current += 2;
		return;
	}

	if(StringAt((m_current - 1), 3, "ACS", "")
		AND ((m_current + 1) == m_last)
		AND !StringAt((m_current - 4), 6, "ISAACS", ""))
	{
		MetaphAdd("X");
		m_current += 2;
		return;
	}

	//else
	if(!StringAt((m_current - 1), 1, "C", "K", "G", "Q", ""))
	{
		MetaphAdd("K");
	}

	//name sent in 'mac caffrey', 'mac gregor
	if(StringAt((m_current + 1), 2, " C", " Q", " G", ""))
	{
		m_current += 2;
	}
	else
	{
		if(StringAt((m_current + 1), 1, "C", "K", "Q", "")
			AND !StringAt((m_current + 1), 2, "CE", "CI", ""))
		{
			m_current += 2;
			// account for combinations such as Ro-ckc-liffe
			if(StringAt((m_current), 1, "C", "K", "Q", "")
				AND !StringAt((m_current + 1), 2, "CE", "CI", ""))
			{
				m_current++;
			}
		}
		else
		{
			m_current++;
		}
	}
}

/**
 * Encodes cases where 'C' is silent at beginning of word
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_Silent_C_At_Beginning()
{
    //skip these when at start of word
    if((m_current == 0)
		AND StringAt(m_current, 2, "CT", "CN", ""))
	{
        m_current += 1;
		return true;
	}

	return false;
}

/**
 * Encodes exceptions where "-CA-" should encode to S
 * instead of K including cases where the cedilla has not been used
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_CA_To_S()
{
	// Special case: 'caesar'.
	// Also, where cedilla not used, as in "linguica" => LNKS
	if(((m_current == 0) AND StringAt(m_current, 4, "CAES", "CAEC", "CAEM", ""))
		OR StringAt(0, 8, "FRANCAIS", "FRANCAIX",  "LINGUICA", "")
		OR StringAt(0, 6, "FACADE", "")
		OR StringAt(0, 9, "GONCALVES", "PROVENCAL", ""))
	{
		MetaphAdd("S");
		AdvanceCounter(2, 1);
		return true;
	}

	return false;
}

/**
 * Encodes exceptions where "-CO-" encodes to S instead of K
 * including cases where the cedilla has not been used
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_CO_To_S()
{
	// e.g. 'coelecanth' => SLKN0
	if((StringAt(m_current, 4, "COEL", "")
			AND (IsVowel(m_current + 4) OR ((m_current + 3) == m_last)))
		OR StringAt(m_current, 5, "COENA", "COENO", "")
		OR StringAt(0, 8, "FRANCOIS", "MELANCON", "")
		OR StringAt(0, 6, "GARCON", ""))
	{
		MetaphAdd("S");
		AdvanceCounter(3, 1);
		return true;
	}

	return false;
}

/**
 * Encode "-CH-"
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_CH()
{
	if(StringAt(m_current, 2, "CH", ""))
	{
		if(Encode_CHAE()
			OR Encode_CH_To_H()
			OR Encode_Silent_CH()
			OR Encode_ARCH()
			// Encode_CH_To_X() should be
			// called before the germanic
			// and greek encoding functions
			OR Encode_CH_To_X()
			OR Encode_English_CH_To_K()
			OR Encode_Germanic_CH_To_K()
			OR Encode_Greek_CH_Initial()
			OR Encode_Greek_CH_Non_Initial())
		{
			return true;
		}

		if(m_current > 0)
		{
			if(StringAt(0, 2, "MC", "")
						AND (m_current == 1))
			{
				//e.g., "McHugh"
				MetaphAdd("K");
			}
			else
			{
				MetaphAdd("X", "K");
			}
		}
		else
		{
			MetaphAdd("X");
		}

		m_current += 2;
		return true;
	}

	return false;
}

/**
 * Encodes "-CHAE-"
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_CHAE()
{
	// e.g. 'michael'
	if(((m_current > 0) AND StringAt((m_current + 2), 2, "AE", "")))
	{
		if(StringAt(0, 7, "RACHAEL", ""))
		{
			MetaphAdd("X");
		}
		else if(!StringAt((m_current - 1), 1, "C", "K", "G", "Q", ""))
		{
			MetaphAdd("K");
		}

		AdvanceCounter(4, 2);
		return true;
	}

	return false;
}

/**
 * Encdoes transliterations from the hebrew where the
 * sound 'kh' is represented as "-CH-". The normal pronounciation
 * of this in english is either 'h' or 'kh', and alternate
 * spellings most often use "-H-"
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_CH_To_H()
{
	// hebrew => 'H', e.g. 'channukah', 'chabad'
	if(((m_current == 0)
		AND (StringAt((m_current + 2), 3, "AIM", "ETH", "ELM", "")
		OR StringAt((m_current + 2), 4, "ASID", "AZAN", "")
		OR StringAt((m_current + 2), 5, "UPPAH", "UTZPA", "ALLAH", "ALUTZ", "AMETZ", "")
		OR StringAt((m_current + 2), 6, "ESHVAN", "ADARIM", "ANUKAH", "")
		OR StringAt((m_current + 2), 7, "ALLLOTH", "ANNUKAH", "AROSETH", "")))
		// and an irish name with the same encoding
		OR StringAt((m_current - 3), 7, "CLACHAN", ""))
	{
		MetaphAdd("H");
		AdvanceCounter(3, 2);
		return true;
	}

	return false;
}

/**
 * Encodes cases where "-CH-" is not pronounced
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_Silent_CH()
{
	// '-ch-' not pronounced
	if(StringAt((m_current - 2), 7, "FUCHSIA", "")
		OR StringAt((m_current - 2), 5, "YACHT", "")
		OR StringAt(0, 8, "STRACHAN", "")
		OR StringAt(0, 8, "CRICHTON", "")
		OR (StringAt((m_current - 3), 6, "DRACHM", "")
			AND !StringAt((m_current - 3), 7, "DRACHMA", "")))
	{
		m_current += 2;
		return true;
	}

	return false;
}

/**
 * Encodes "-CH-" to X
 * English language patterns
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_CH_To_X()
{
	// e.g. 'approach', 'beach'
	if((StringAt((m_current - 2), 4, "OACH", "EACH", "EECH", "OUCH", "OOCH", "MUCH", "SUCH", "")
			AND !StringAt((m_current - 3), 5, "JOACH", ""))
		// e.g. 'dacha', 'macho'
		OR (((m_current + 2) == m_last ) AND StringAt((m_current - 1), 4, "ACHA", "ACHO", ""))
		OR (StringAt(m_current, 4, "CHOT", "CHOD", "CHAT", "") AND ((m_current + 3) == m_last))
		OR ((StringAt((m_current - 1), 4, "OCHE", "") AND ((m_current + 2) == m_last))
				AND !StringAt((m_current - 2), 5, "DOCHE", ""))
		OR StringAt((m_current - 4), 6, "ATTACH", "DETACH", "KOVACH", "")
		OR StringAt((m_current - 5), 7, "SPINACH", "")
		OR StringAt(0, 6, "MACHAU", "")
		OR StringAt((m_current - 4), 8, "PARACHUT", "")
		OR StringAt((m_current - 5), 8, "MASSACHU", "")
		OR (StringAt((m_current - 3), 5, "THACH", "") AND !StringAt((m_current - 1), 4, "ACHE", ""))
		OR StringAt((m_current - 2), 6, "VACHON", "") )
	{
		MetaphAdd("X");
		m_current += 2;
		return true;
	}

	return false;
}

/**
 * Encodes "-CH-" to K in contexts of
 * initial "A" or "E" follwed by "CH"
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_English_CH_To_K()
{
	//'ache', 'echo', alternate spelling of 'michael'
	if(((m_current == 1) AND RootOrInflections(m_inWord, "ACHE"))
		OR (((m_current > 3) AND RootOrInflections(&m_inWord[m_current - 1], "ACHE"))
			AND (StringAt(0, 3, "EAR", "")
				OR StringAt(0, 4, "HEAD", "BACK", "")
				OR StringAt(0, 5, "HEART", "BELLY", "TOOTH", "")))
		OR StringAt((m_current - 1), 4, "ECHO", "")
		OR StringAt((m_current - 2), 7, "MICHEAL", "")
		OR StringAt((m_current - 4), 7, "JERICHO", "")
		OR StringAt((m_current - 5), 7, "LEPRECH", ""))
	{
		MetaphAdd("K", "X");
		m_current += 2;
		return true;
	}

	return false;
}

/**
 * Encodes "-CH-" to K in mostly germanic context
 * of internal "-ACH-", with exceptions
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_Germanic_CH_To_K()
{
	// various germanic
	// "<consonant><vowel>CH-"implies a german word where 'ch' => K
	if(((m_current > 1)
		AND !IsVowel(m_current - 2)
		AND StringAt((m_current - 1), 3, "ACH", "")
		AND !StringAt((m_current - 2), 7, "MACHADO", "MACHUCA", "LACHANC", "LACHAPE", "KACHATU", "")
		AND !StringAt((m_current - 3), 7, "KHACHAT", "")
		AND ((m_inWord[m_current + 2] != 'I')
			AND ((m_inWord[m_current + 2] != 'E')
				OR StringAt((m_current - 2), 6, "BACHER", "MACHER", "MACHEN", "LACHER", "")) )
		// e.g. 'brecht', 'fuchs'
		OR (StringAt((m_current + 2), 1, "T", "S", "") AND !(StringAt(0, 11, "WHICHSOEVER", "")
															OR StringAt(0, 9, "LUNCHTIME", "") ))
		// e.g. 'andromache'
		OR StringAt(0, 4, "SCHR", "")
		OR ((m_current > 2) AND StringAt((m_current - 2), 5, "MACHE", ""))
		OR ((m_current == 2) AND StringAt((m_current - 2), 4, "ZACH", ""))
		OR StringAt((m_current - 4), 6, "SCHACH", "")
		OR StringAt((m_current - 1), 5, "ACHEN", "")
		OR StringAt((m_current - 3), 5, "SPICH", "ZURCH", "BUECH", "")
		OR (StringAt((m_current - 3), 5, "KIRCH", "JOACH", "BLECH", "MALCH", "")
			// "kirch" and "blech" => 'X'
			AND !(StringAt((m_current - 3), 8, "KIRCHNER", "") OR ((m_current + 1) == m_last)))
		OR (((m_current + 1) == m_last) && StringAt((m_current - 2), 4, "NICH", "LICH", "BACH", ""))
		OR (((m_current + 1) == m_last)
			AND StringAt((m_current - 3), 5, "URICH", "BRICH", "ERICH", "DRICH", "NRICH", "")
			AND !StringAt((m_current - 5), 7, "ALDRICH", "")
			AND !StringAt((m_current - 6), 8, "GOODRICH", "")
			AND !StringAt((m_current - 7), 9, "GINGERICH", "")))
		OR (((m_current + 1) == m_last) && StringAt((m_current - 4), 6, "ULRICH", "LFRICH", "LLRICH",
																		"EMRICH", "ZURICH", "EYRICH", ""))
		// e.g., 'wachtler', 'wechsler', but not 'tichner'
		OR ((StringAt((m_current - 1), 1, "A", "O", "U", "E", "") OR (m_current == 0))
					AND StringAt((m_current + 2), 1, "L", "R", "N", "M", "B", "H", "F", "V", "W", " ", "")))
	{
		// "CHR/L-" e.g. 'chris' do not get
		// alt pronunciation of 'X'
		if(StringAt((m_current + 2), 1, "R", "L", ""))
		{
			MetaphAdd("K");
		}
		else
		{
			MetaphAdd("K", "X");
		}
		m_current += 2;
		return true;
	}

	return false;
}

/**
 * Encode "-ARCH-". Some occurances are from greek roots and therefore encode
 * to 'K', others are from english words and therefore encode to 'X'
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_ARCH()
{
	if(StringAt((m_current - 2), 4, "ARCH", ""))
	{
		// "-ARCH-" has many combining forms where "-CH-" => K because of its
		// derivation from the greek
		if(((IsVowel(m_current + 2) AND StringAt((m_current - 2), 5, "ARCHA", "ARCHI", "ARCHO", "ARCHU", "ARCHY", ""))
			OR StringAt((m_current - 2), 6, "ARCHEA", "ARCHEG", "ARCHEO", "ARCHET", "ARCHEL", "ARCHES", "ARCHEP",
											"ARCHEM", "ARCHEN", "")
			OR (StringAt((m_current - 2), 4, "ARCH", "") AND (((m_current + 1) == m_last)))
			OR StringAt(0, 7, "MENARCH", ""))
			AND (!RootOrInflections(m_inWord, "ARCH")
				AND !StringAt((m_current - 4), 6, "SEARCH", "POARCH", "")
				AND !StringAt(0, 9, "ARCHENEMY", "ARCHIBALD", "ARCHULETA", "ARCHAMBAU", "")
				AND !StringAt(0, 6, "ARCHER", "ARCHIE", "")
				AND !((((StringAt((m_current - 3), 5, "LARCH", "MARCH", "PARCH", "")
						OR StringAt((m_current - 4), 6, "STARCH", ""))
						AND !(StringAt(0, 6, "EPARCH", "")
								OR StringAt(0, 7, "NOMARCH", "")
								OR StringAt(0, 8, "EXILARCH", "HIPPARCH", "MARCHESE", "")
								OR StringAt(0, 9, "ARISTARCH", "")
								OR StringAt(0, 9, "MARCHETTI", "")) )
						OR RootOrInflections(m_inWord, "STARCH"))
						AND (!StringAt((m_current - 2), 5, "ARCHU", "ARCHY", "")
								OR StringAt(0, 7, "STARCHY", "")))))
		{
			MetaphAdd("K", "X");
		}
		else
		{
			MetaphAdd("X");
		}
		m_current += 2;
		return true;
	}

	return false;
}

/**
 * Encode "-CH-" to K when from greek roots
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_Greek_CH_Initial()
{
	// greek roots e.g. 'chemistry', 'chorus', ch at beginning of root
	if((StringAt(m_current, 6, "CHAMOM", "CHARAC", "CHARIS", "CHARTO", "CHARTU", "CHARYB", "CHRIST", "CHEMIC", "CHILIA", "")
		OR (StringAt(m_current, 5, "CHEMI", "CHEMO", "CHEMU", "CHEMY", "CHOND", "CHONA", "CHONI", "CHOIR", "CHASM",
								   "CHARO", "CHROM", "CHROI", "CHAMA", "CHALC", "CHALD", "CHAET","CHIRO", "CHILO",
								   "CHELA", "CHOUS", "CHEIL", "CHEIR", "CHEIM", "CHITI", "CHEOP", "")
				AND !(StringAt(m_current, 6, "CHEMIN", "") OR StringAt((m_current - 2), 8, "ANCHONDO", "")))
		OR (StringAt(m_current, 5, "CHISM", "CHELI", "")
		// exclude spanish "machismo"
			AND !(StringAt(0, 8, "MACHISMO", "")
			// exclude some french words
				OR StringAt(0, 10, "REVANCHISM", "")
				OR StringAt(0, 9, "RICHELIEU", "")
				OR (StringAt(0, 5, "CHISM", "") AND (m_length == 5))
				OR StringAt(0, 6, "MICHEL", "")))
		// include e.g. "chorus", "chyme", "chaos"
		OR (StringAt(m_current, 4, "CHOR", "CHOL", "CHYM", "CHYL", "CHLO", "CHOS", "CHUS", "CHOE", "")
					AND !StringAt(0, 6, "CHOLLO", "CHOLLA", "CHORIZ", ""))
		// "chaos" => K but not "chao"
		OR (StringAt(m_current, 4, "CHAO", "") && ((m_current + 3) != m_last))
		// e.g. "abranchiate"
		OR (StringAt(m_current, 4, "CHIA", "")  AND !(StringAt(0, 10, "APPALACHIA", "") || StringAt(0, 7, "CHIAPAS", "")))
		// e.g. "chimera"
		OR StringAt(m_current, 7, "CHIMERA", "CHIMAER", "CHIMERI", "")
		// e.g. "chameleon"
		OR ((m_current == 0) AND StringAt(m_current, 5, "CHAME", "CHELO", "CHITO", "") )
		// e.g. "spirochete"
		OR ((((m_current + 4) == m_last) OR ((m_current + 5) == m_last)) AND StringAt((m_current - 1), 6, "OCHETE", "")))
		// more exceptions where "-CH-" => X e.g. "chortle", "crocheter"
			AND !((StringAt(0, 5, "CHORE",  "CHOLO", "CHOLA", "") AND (m_length == 5))
				OR StringAt(m_current, 5, "CHORT", "CHOSE", "")
				OR StringAt((m_current - 3), 7, "CROCHET", "")
				OR StringAt(0, 7, "CHEMISE", "CHARISE", "CHARISS", "CHAROLE", "")) )
	{
		// "CHR/L-" e.g. 'christ', 'chlorine' do not get
		// alt pronunciation of 'X'
		if(StringAt((m_current + 2), 1, "R", "L", "")
			OR SlavoGermanic())
		{
			MetaphAdd("K");
		}
		else
		{
			MetaphAdd("K", "X");
		}
		m_current += 2;
		return true;
	}

	return false;
}

/**
 * Encode a variety of greek and some german roots where "-CH-" => K
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_Greek_CH_Non_Initial()
{
	//greek & other roots e.g. 'tachometer', 'orchid', ch in middle or end of root
	if(StringAt((m_current - 2), 6, "ORCHID", "NICHOL", "MECHAN", "LICHEN", "MACHIC", "PACHEL",
									"RACHIF", "RACHID", "RACHIS", "RACHIC", "MICHAL", "")
		OR StringAt((m_current - 3), 5, "MELCH", "GLOCH", "TRACH", "TROCH", "BRACH", "SYNCH", "PSYCH",
										"STICH", "PULCH", "EPOCH", "")
		OR (StringAt((m_current - 3), 5, "TRICH", "") AND !StringAt((m_current - 5), 7, "OSTRICH", ""))
		OR (StringAt((m_current - 2), 4, "TYCH", "TOCH", "BUCH", "MOCH", "CICH", "DICH", "NUCH", "EICH", "LOCH",
										 "DOCH", "ZECH", "WYCH", "")
			AND !(StringAt((m_current - 4), 9, "INDOCHINA", "")OR StringAt((m_current - 2), 6, "BUCHON", "")))
		OR StringAt((m_current - 2), 5, "LYCHN", "TACHO", "ORCHO", "ORCHI", "LICHO", "")
		OR (StringAt((m_current - 1), 5, "OCHER", "ECHIN", "ECHID", "") AND ((m_current == 1) OR (m_current == 2)))
		OR StringAt((m_current - 4), 6, "BRONCH", "STOICH", "STRYCH", "TELECH", "PLANCH", "CATECH", "MANICH",
										"MALACH", "BIANCH", "DIDACH", "")
		OR (StringAt((m_current - 1), 4, "ICHA", "ICHN","") AND (m_current == 1))
		OR StringAt((m_current - 2), 8, "ORCHESTR", "")
		OR StringAt((m_current - 4), 8, "BRANCHIO", "BRANCHIF", "")
		OR (StringAt((m_current - 1), 5, "ACHAB", "ACHAD", "ACHAN", "ACHAZ", "")
			AND !StringAt((m_current - 2), 7, "MACHADO", "LACHANC", ""))
		OR StringAt((m_current - 1), 6, "ACHISH", "ACHILL", "ACHAIA", "ACHENE", "")
		OR StringAt((m_current - 1), 7, "ACHAIAN", "ACHATES", "ACHIRAL", "ACHERON", "")
		OR StringAt((m_current - 1), 8, "ACHILLEA", "ACHIMAAS", "ACHILARY", "ACHELOUS", "ACHENIAL", "ACHERNAR", "")
		OR StringAt((m_current - 1), 9, "ACHALASIA", "ACHILLEAN", "ACHIMENES", "")
		OR StringAt((m_current - 1), 10, "ACHIMELECH", "ACHITOPHEL", "")
		// e.g. 'inchoate'
		OR (((m_current - 2) == 0) AND (StringAt((m_current - 2), 6, "INCHOA", "")
		// e.g. 'ischemia'
		OR StringAt(0, 4, "ISCH", "")) )
		// e.g. 'ablimelech', 'antioch', 'pentateuch'
		OR (((m_current + 1) == m_last) AND StringAt((m_current - 1), 1, "A", "O", "U", "E", "")
			AND !(StringAt(0, 7, "DEBAUCH", "")
					OR StringAt((m_current - 2), 4, "MUCH", "SUCH", "KOCH", "")
					OR StringAt((m_current - 5), 7, "OODRICH", "ALDRICH", ""))))
	{
		MetaphAdd("K", "X");
		m_current += 2;
		return true;
	}

	return false;
}

/**
 * Encodes reliably italian "-CCIA-"
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_CCIA()
{
	//e.g., 'focaccia'
	if(StringAt((m_current + 1), 3, "CIA", ""))
	{
		MetaphAdd("X", "S");
		m_current += 2;
		return true;
	}

	return false;
}

/**
 * Encode "-CC-"
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_CC()
{
	//double 'C', but not if e.g. 'McClellan'
	if(StringAt(m_current, 2, "CC", "") AND !((m_current == 1) AND (m_inWord[0] == 'M')))
	{
		// exception
		if (StringAt((m_current - 3), 7, "FLACCID", ""))
		{
			MetaphAdd("S");
			AdvanceCounter(3, 2);
			return true;
		}

		//'bacci', 'bertucci', other italian
		if((((m_current + 2) == m_last) AND StringAt((m_current + 2), 1, "I", ""))
			OR StringAt((m_current + 2), 2, "IO", "")
			OR (((m_current + 4) == m_last) AND StringAt((m_current + 2), 3, "INO", "INI", "")))
		{
			MetaphAdd("X");
			AdvanceCounter(3, 2);
			return true;
		}

		//'accident', 'accede' 'succeed'
		if(StringAt((m_current + 2), 1, "I", "E", "Y", "")
			//except 'bellocchio','bacchus', 'soccer' get K
			AND !((m_inWord[m_current + 2] == 'H')
				OR StringAt((m_current - 2), 6, "SOCCER", "")))
		{
			MetaphAdd("KS");
			AdvanceCounter(3, 2);
			return true;

		}
		else
		{
			//Pierce's rule
			MetaphAdd("K");
			m_current += 2;
			return true;
		}
	}

	return false;
}

/**
 * Encode cases where the consonant following "C" is redundant
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_CK_CG_CQ()
{
	if(StringAt(m_current, 2, "CK", "CG", "CQ", ""))
	{
		// eastern european spelling e.g. 'gorecki' == 'goresky'
		if(StringAt(m_current, 3, "CKI", "CKY", "")
			AND ((m_current + 2) == m_last)
			AND (m_length > 6))
		{
			MetaphAdd("K", "SK");
		}
		else
		{
			MetaphAdd("K");
		}
		m_current += 2;

		if(StringAt(m_current, 1, "K", "G", "Q", ""))
		{
			m_current++;
		}
		return true;
	}

	return false;
}

/**
 * Encode cases where "C" preceeds a front vowel such as "E", "I", or "Y".
 * These cases most likely => S or X
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_C_Front_Vowel()
{
	if(StringAt(m_current, 2, "CI", "CE", "CY", ""))
	{
		if(Encode_British_Silent_CE()
			OR Encode_CE()
			OR Encode_CI()
			OR Encode_Latinate_Suffixes())
		{
			AdvanceCounter(2, 1);
			return true;
		}

		MetaphAdd("S");
		AdvanceCounter(2, 1);
		return true;
	}

	return false;
}

/**
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_British_Silent_CE()
{
	// english place names like e.g.'gloucester' pronounced glo-ster
	if((StringAt((m_current + 1), 5, "ESTER", "") AND ((m_current + 5) == m_last))
		OR StringAt((m_current + 1), 10, "ESTERSHIRE", ""))
	{
		return true;
	}

	return false;
}

/**
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_CE()
{
	// 'ocean', 'commercial', 'provincial', 'cello', 'fettucini', 'medici'
	if((StringAt((m_current + 1), 3, "EAN", "") AND IsVowel(m_current - 1))
		// e.g. 'rosacea'
		OR (StringAt((m_current - 1), 4, "ACEA", "")
			AND ((m_current + 2) == m_last)
			AND !StringAt(0, 7, "PANACEA", ""))
		// e.g. 'botticelli', 'concerto'
		OR StringAt((m_current + 1), 4, "ELLI", "ERTO", "EORL", "")
		// some italian names familiar to americans
		OR (StringAt((m_current - 3), 5, "CROCE", "") AND ((m_current + 1) == m_last))
		OR StringAt((m_current - 3), 5, "DOLCE", "")
		OR StringAt((m_current - 5), 7, "VERSACE", "")
		// e.g. 'cello'
		OR (StringAt((m_current + 1), 4, "ELLO", "")
			AND ((m_current + 4) == m_last)))
	{
		MetaphAdd("X", "S");
		return true;
	}

	return false;
}

/**
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_CI()
{
	// with consonant before C
	// e.g. 'fettucini', but exception for the americanized pronunciation of 'mancini'
	if(((StringAt((m_current + 1), 3, "INI", "") AND !StringAt(0, 7, "MANCINI", "")) AND ((m_current + 3) == m_last))
		// e.g. 'medici'
		OR (StringAt((m_current - 1), 3, "ICI", "") AND ((m_current + 1) == m_last))
		// e.g. "commercial', 'provincial', 'cistercian'
		OR StringAt((m_current - 1), 5, "RCIAL", "NCIAL", "RCIAN", "UCIUS", "")
		// special cases
		OR StringAt((m_current - 3), 6, "MARCIA", "")
		OR StringAt((m_current - 2), 7, "ANCIENT", ""))
	{
		MetaphAdd("X", "S");
		return true;
	}

	// with vowel before C
	if(((StringAt(m_current, 3, "CIO", "CIE", "CIA", "")
		AND IsVowel(m_current - 1))
		// e.g. "ciao"
		OR StringAt((m_current + 1), 3, "IAO", ""))
		AND !StringAt((m_current - 4), 8, "COERCION", ""))
	{
		if((StringAt(m_current, 4, "CIAN", "CIAL", "CIAO", "CIES", "CIOL", "CION", "")
			// exception - "glacier" => 'X' but "spacier" = > 'S'
			OR StringAt((m_current - 3), 7, "GLACIER", "")
			OR StringAt(m_current, 5, "CIENT", "CIENC", "CIOUS", "CIATE", "CIATI", "CIATO", "CIABL", "CIARY", "")
			OR (((m_current + 2) == m_last) AND StringAt(m_current, 3, "CIA", "CIO", ""))
			OR (((m_current + 3) == m_last) AND StringAt(m_current, 3, "CIAS", "CIOS", "")))
			// exceptions
			AND !(StringAt((m_current - 4), 11, "ASSOCIATION", "")
				OR StringAt(0, 4, "OCIE", "")
				// exceptions mostly because these names are usually from
				// the spanish rather than the italian in america
				OR StringAt((m_current - 2), 5, "LUCIO", "")
				OR StringAt((m_current - 2), 6, "MACIAS", "")
				OR StringAt((m_current - 3), 6, "GRACIE", "GRACIA", "")
				OR StringAt((m_current - 2), 7, "LUCIANO", "")
				OR StringAt((m_current - 3), 8, "MARCIANO", "")
				OR StringAt((m_current - 4), 7, "PALACIO", "")
				OR StringAt((m_current - 4), 9, "FELICIANO", "")
				OR StringAt((m_current - 5), 8, "MAURICIO", "")
				OR StringAt((m_current - 7), 11, "ENCARNACION", "")
				OR StringAt((m_current - 4), 8, "POLICIES", "")
				OR StringAt((m_current - 2), 8, "HACIENDA", "")
				OR StringAt((m_current - 6), 9, "ANDALUCIA", "")
				OR StringAt((m_current - 2), 5, "SOCIO", "SOCIE", "")))
		{
			MetaphAdd("X", "S");
		}
		else
		{
			MetaphAdd("S", "X");
		}

		return true;
	}

	// exception
	if(StringAt((m_current - 4), 8, "COERCION", ""))
	{
		MetaphAdd("J");
		return true;
	}

	return false;
}

/**
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_Latinate_Suffixes()
{
	if(StringAt((m_current + 1), 4, "EOUS", "IOUS", ""))
	{
		MetaphAdd("X", "S");
		return true;
	}

	return false;
}

/**
 * Encodes some exceptions where "C" is silent
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_Silent_C()
{
	if(StringAt((m_current + 1), 1, "T", "S", ""))
	{
		if (StringAt(0, 11, "CONNECTICUT", "")
			OR StringAt(0, 6, "INDICT", "TUCSON", ""))
		{
			m_current++;
			return true;
		}
	}

	return false;
}

/**
 * Encodes slavic spellings or transliterations
 * written as "-CZ-"
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_CZ()
{
	if(StringAt((m_current + 1), 1, "Z", "")
		AND !StringAt((m_current - 1), 6, "ECZEMA", ""))
	{
		if(StringAt(m_current, 4, "CZAR", ""))
		{
			MetaphAdd("S");
		}
		// otherwise most likely a czech word...
		else
		{
			MetaphAdd("X");
		}
		m_current += 2;
		return true;
	}

	return false;
}

/**
 * "-CS" special cases
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_CS()
{
	// give an 'etymological' 2nd
	// encoding for "kovacs" so
	// that it matches "kovach"
	if(StringAt(0, 6, "KOVACS", ""))
	{
		MetaphAdd("KS", "X");
		m_current += 2;
		return true;
	}

	if(StringAt((m_current - 1), 3, "ACS", "")
		AND ((m_current + 1) == m_last)
		AND !StringAt((m_current - 4), 6, "ISAACS", ""))
	{
		MetaphAdd("X");
		m_current += 2;
		return true;
	}

	return false;
}

/**
 * Encode "-D-"
 *
 */
void Metaphone3::Encode_D()
{
	if(Encode_DG()
		OR Encode_DJ()
		OR Encode_DT_DD()
		OR Encode_D_To_J()
		OR Encode_DOUS()
		OR Encode_Silent_D())
	{
		return;
	}

	if(m_encodeExact)
	{
		// "final de-voicing" in this case
		// e.g. 'missed' == 'mist'
		if((m_current == m_last)
			AND StringAt((m_current - 3), 4, "SSED", ""))
		{
			MetaphAdd("T");
		}
		else
		{
			MetaphAdd("D");
		}
	}
	else
	{
		MetaphAdd("T");
	}
	m_current++;
}

/**
 * Encode "-DG-"
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_DG()
{
	if(StringAt(m_current, 2, "DG", ""))
	{
		// excludes exceptions e.g. 'edgar',
		// or cases where 'g' is first letter of combining form
		// e.g. 'handgun', 'waldglas'
		if(StringAt((m_current + 2), 1, "A", "O", "")
			// e.g. "midgut"
			OR StringAt((m_current + 1), 3, "GUN", "GUT", "")
			// e.g. "handgrip"
			OR StringAt((m_current + 1), 4, "GEAR", "GLAS", "GRIP", "GREN", "GILL", "GRAF", "")
			// e.g. "mudgard"
			OR StringAt((m_current + 1), 5, "GUARD", "GUILT", "GRAVE", "GRASS", "")
			// e.g. "woodgrouse"
			OR StringAt((m_current + 1), 6, "GROUSE", ""))
		{
			MetaphAddExactApprox("DG", "TK");
		}
		else
		{
			//e.g. "edge", "abridgment"
			MetaphAdd("J");
		}
		m_current += 2;
		return true;
	}

	return false;
}

/**
 * Encode "-DJ-"
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_DJ()
{
	// e.g. "adjacent"
	if(StringAt(m_current, 2, "DJ", ""))
	{
		MetaphAdd("J");
		m_current += 2;
		return true;
	}

	return false;
}

/**
 * Encode "-DD-" and "-DT-"
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_DT_DD()
{
	// eat redundant 'T' or 'D'
	if(StringAt(m_current, 2, "DT", "DD", ""))
	{
		if(StringAt(m_current, 3, "DTH",  ""))
		{
			MetaphAddExactApprox("D0", "T0");
			m_current += 3;
		}
		else
		{
			if(m_encodeExact)
			{
				// devoice it
				if(StringAt(m_current, 2, "DT", ""))
				{
					MetaphAdd("T");
				}
				else
				{
					MetaphAdd("D");
				}
			}
			else
			{
				MetaphAdd("T");
			}
			m_current += 2;
		}
		return true;
	}

	return false;
}

/**
 * Encode cases where "-DU-" "-DI-", and "-DI-" => J
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_D_To_J()
{
	// e.g. "module", "adulate"
	if((StringAt(m_current, 3, "DUL", "")
			AND (IsVowel(m_current - 1) AND IsVowel(m_current + 3)))
		// e.g. "soldier", "grandeur", "procedure"
		OR (((m_current + 3) == m_last)
			AND StringAt((m_current - 1) , 5, "LDIER", "NDEUR", "EDURE", "RDURE", ""))
		OR StringAt((m_current - 3), 7, "CORDIAL", "")
		// e.g.  "pendulum", "education"
		OR StringAt((m_current - 1), 5, "NDULA", "NDULU", "EDUCA", "")
		// e.g. "individual", "individual", "residuum"
		OR StringAt((m_current - 1), 4, "ADUA", "IDUA", "IDUU", ""))
	{
		MetaphAddExactApprox("J", "D", "J", "T");
		AdvanceCounter(2, 1);
		return true;
	}

	return false;
}

/**
 * Encode latinate suffix "-DOUS" where 'D' is pronounced as J
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_DOUS()
{
	// e.g. "assiduous", "arduous"
	if(StringAt((m_current + 1), 4, "UOUS", ""))
	{
		MetaphAddExactApprox("J", "D", "J", "T");
		AdvanceCounter(4, 1);
		return true;
	}

	return false;
}

/**
 * Encode silent "-D-"
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_Silent_D()
{
	// silent 'D' e.g. 'wednesday', 'handsome'
	if(StringAt((m_current - 2), 9, "WEDNESDAY", "")
		OR StringAt((m_current - 3), 7, "HANDKER", "HANDSOM", "WINDSOR", "")
		// french silent D at end in words or names familiar to americans
		OR StringAt((m_current - 5), 6, "PERNOD", "ARTAUD", "RENAUD", "")
		OR StringAt((m_current - 6), 7, "RIMBAUD", "MICHAUD", "BICHAUD", ""))
	{
		m_current++;
		return true;
	}

	return false;
}

/**
 * Encode "-F-"
 *
 */
void Metaphone3::Encode_F()
{
	// Encode cases where "-FT-" => "T" is usually silent
	// e.g. 'often', 'soften'
	// This should really be covered under "T"!
	if(StringAt((m_current - 1), 5, "OFTEN", ""))
	{
		MetaphAdd("F", "FT");
		m_current += 2;
		return;
	}

	// eat redundant 'F'
	if(m_inWord[m_current + 1] == 'F')
	{
		m_current += 2;
	}
	else
	{
		m_current++;
	}

	MetaphAdd("F");

}

/**
 * Encode "-G-"
 *
 */
void Metaphone3::Encode_G()
{
	if(Encode_Silent_G_At_Beginning()
		OR Encode_GG()
		OR Encode_GK()
		OR Encode_GH()
		OR Encode_Silent_G()
		OR Encode_GN()
		OR Encode_GL()
		OR Encode_Initial_G_Front_Vowel()
		OR Encode_NGER()
		OR Encode_GER()
		OR Encode_GEL()
		OR Encode_Non_Initial_G_Front_Vowel()
		OR Encode_GA_To_J())
	{
		return;
	}

	if(!StringAt((m_current - 1), 1, "C", "K", "G", "Q", ""))
	{
		MetaphAddExactApprox("G", "K");
	}

	m_current++;
}

/**
 * Encode cases where 'G' is silent at beginning of word
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_Silent_G_At_Beginning()
{
	//skip these when at start of word
    if((m_current == 0)
		AND StringAt(m_current, 2, "GN", ""))
	{
        m_current += 1;
		return true;
	}

	return false;
}

/**
 * Encode "-GG-"
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_GG()
{
	if(m_inWord[m_current + 1] == 'G')
	{
		// italian e.g, 'loggia', 'caraveggio', also 'suggest' and 'exaggerate'
		if(StringAt((m_current - 1), 5, "AGGIA", "OGGIA", "AGGIO", "EGGIO", "EGGIA", "IGGIO", "")
			// 'ruggiero' but not 'snuggies'
			OR (StringAt((m_current - 1), 5, "UGGIE", "")
					AND !(((m_current + 3) == m_last) OR ((m_current + 4) == m_last)))
			OR (((m_current + 2) == m_last) AND StringAt((m_current - 1), 4, "AGGI", "OGGI", ""))
			OR StringAt((m_current - 2), 6, "SUGGES", "XAGGER", "REGGIE", ""))
		{
			// expection where "-GG-" => KJ
			if (StringAt((m_current - 2), 7, "SUGGEST", ""))
			{
				MetaphAddExactApprox("G", "K");
			}

			MetaphAdd("J");
			AdvanceCounter(3, 2);
		}
		else
		{
			MetaphAddExactApprox("G", "K");
			m_current += 2;
		}
		return true;
	}

	return false;
}

/**
 * Encode "-GK-"
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_GK()
{
	// 'gingko'
	if(m_inWord[m_current + 1] == 'K')
	{
		MetaphAdd("K");
		m_current += 2;
		return true;
	}

	return false;
}

/**
 * Encode "-GH-"
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_GH()
{
	if(m_inWord[m_current + 1] == 'H')
	{
		if(Encode_GH_After_Consonant()
			OR Encode_Initial_GH()
			OR Encode_GH_To_J()
			OR Encode_GH_To_H()
			OR Encode_UGHT()
			OR Encode_GH_H_Part_Of_Other_Word()
			OR Encode_Silent_GH()
			OR Encode_GH_To_F())
		{
			return true;
		}

		// default
		MetaphAddExactApprox("G", "K");
		m_current += 2;
		return true;
	}

	return false;
}

/**
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_GH_After_Consonant()
{
	// e.g. 'burgher', 'bingham'
		if((m_current > 0)
			AND !IsVowel(m_current - 1)
			// not e.g. 'greenhalgh'
			AND !(StringAt((m_current - 3), 5, "HALGH", "")
					AND ((m_current + 1) == m_last)))
	{
		MetaphAddExactApprox("G", "K");
		m_current += 2;
		return true;
	}

	return false;
}

/**
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_Initial_GH()
{
	if(m_current < 3)
	{
		// e.g. "ghislane", "ghiradelli"
		if(m_current == 0)
		{
			if(m_inWord[m_current + 2] == 'I')
			{
				MetaphAdd("J");
			}
			else
			{
				MetaphAddExactApprox("G", "K");
			}
			m_current += 2;
			return true;
		}
	}

	return false;
}

/**
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_GH_To_J()
{
	// e.g., 'greenhalgh', 'dunkenhalgh', english names
	if(StringAt((m_current - 2), 4, "ALGH", "") AND ((m_current + 1) == m_last))
	{
		MetaphAdd("J", "");
		m_current += 2;
		return true;
	}

	return false;
}
/**
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_GH_To_H()
{
	// special cases
	// e.g., 'donoghue', 'donaghy'
	if((StringAt((m_current - 4), 4, "DONO", "DONA", "") AND IsVowel(m_current + 2))
		OR StringAt((m_current - 5), 9, "CALLAGHAN", ""))
	{
		MetaphAdd("H");
		m_current += 2;
		return true;
	}

	return false;
}

/**
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_UGHT()
{
	//e.g. "ought", "aught", "daughter", "slaughter"
	if(StringAt((m_current - 1), 4, "UGHT", ""))
	{
		if ((StringAt((m_current - 3), 5, "LAUGH", "")
			AND !(StringAt((m_current - 4), 7, "SLAUGHT", "")
				OR StringAt((m_current - 3), 7, "LAUGHTO", "")))
				OR StringAt((m_current - 4), 6, "DRAUGH", ""))
		{
			MetaphAdd("FT");
		}
		else
		{
			MetaphAdd("T");
		}
		m_current += 3;
		return true;
	}

	return false;
}

/**
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_GH_H_Part_Of_Other_Word()
{
	// if the 'H' is the beginning of another word or syllable
	if (StringAt((m_current + 1), 4, "HOUS", "HEAD", "HOLE", "HORN", "HARN", ""))
	{
		MetaphAddExactApprox("G", "K");
		m_current += 2;
		return true;
	}

	return false;
}

/**
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_Silent_GH()
{
	//Parker's rule (with some further refinements) - e.g., 'hugh'
	if(((((m_current > 1) AND StringAt((m_current - 2), 1, "B", "H", "D", "G", "L", "") )
		//e.g., 'bough'
		OR ((m_current > 2)
			AND StringAt((m_current - 3), 1, "B", "H", "D", "K", "W", "N", "P", "V", "")
			AND !StringAt(0, 6, "ENOUGH", ""))
		//e.g., 'broughton'
		OR ((m_current > 3) AND StringAt((m_current - 4), 1, "B", "H", "") )
		//'plough', 'slaugh'
		OR ((m_current > 3) AND StringAt((m_current - 4), 2, "PL", "SL", "") )
		OR ((m_current > 0)
			// 'sigh', 'light'
			AND ((m_inWord[m_current - 1] == 'I')
				OR StringAt(0, 4, "PUGH", "")
				// e.g. 'MCDONAGH', 'MURTAGH', 'CREAGH'
				OR (StringAt((m_current - 1), 3, "AGH", "")
						AND ((m_current + 1) == m_last))
					OR StringAt((m_current - 4), 6, "GERAGH", "DRAUGH", "")
					OR (StringAt((m_current - 3), 5, "GAUGH", "GEOGH", "MAUGH", "")
							AND !StringAt(0, 9, "MCGAUGHEY", ""))
					// exceptions to 'tough', 'rough', 'lough'
					OR (StringAt((m_current - 2), 4, "OUGH", "")
							AND (m_current > 3)
							AND !StringAt((m_current - 4), 6, "CCOUGH", "ENOUGH", "TROUGH", "CLOUGH", "")))))
		// suffixes starting w/ vowel where "-GH-" is usually silent
		AND (StringAt((m_current - 3), 5, "VAUGH", "FEIGH", "LEIGH", "")
			OR StringAt((m_current - 2), 4, "HIGH", "TIGH", "")
			OR ((m_current + 1) == m_last)
			OR (StringAt((m_current + 2), 2, "IE", "EY", "ES", "ER", "ED", "TY", "")
				&& ((m_current + 3) == m_last)
				&& !StringAt((m_current - 5), 9, "GALLAGHER", ""))
			OR (StringAt((m_current + 2), 1, "Y", "") && ((m_current + 2) == m_last))
			OR (StringAt((m_current + 2), 3, "ING", "OUT", "") AND ((m_current + 4) == m_last))
			OR (StringAt((m_current + 2), 4, "ERTY", "") AND ((m_current + 5) == m_last))
			OR (!IsVowel(m_current + 2)
				OR StringAt((m_current - 3), 5, "GAUGH", "GEOGH", "MAUGH", "")
				OR StringAt((m_current - 4), 8, "BROUGHAM", ""))))
		// exceptions where '-g-' pronounced
		AND !(StringAt(0, 6, "BALOGH", "SABAGH", "")
			OR StringAt((m_current - 2), 7, "BAGHDAD", "")
			OR StringAt((m_current - 3), 5, "WHIGH", "")
			OR StringAt((m_current - 5), 7, "SABBAGH", "AKHLAGH", "")))
	{
		// silent - do nothing
		m_current += 2;
		return true;
	}

	return false;
}

/**
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_GH_Special_Cases()
{
	bool handled = false;

	// special case: 'hiccough' == 'hiccup'
	if(StringAt((m_current - 6), 8, "HICCOUGH", ""))
	{
		MetaphAdd("P");
		handled = true;
	}
	// special case: 'lough' alt spelling for scots 'loch'
	else if(StringAt(0, 5, "LOUGH", ""))
	{
		MetaphAdd("K");
		handled = true;
	}
	// hungarian
	else if(StringAt(0, 6, "BALOGH", ""))
	{
		MetaphAddExactApprox("G", "", "K", "");
		handled = true;
	}
	// "maclaughlin"
	else if(StringAt((m_current - 3), 8, "LAUGHLIN", "COUGHLAN", "LOUGHLIN", ""))
	{
		MetaphAdd("K", "F");
		handled = true;
	}
	else if(StringAt((m_current - 3), 5, "GOUGH", "")
			OR StringAt((m_current - 7), 9, "COLCLOUGH", ""))
	{
		MetaphAdd("", "F");
		handled = true;
	}

	if(handled)
	{
		m_current += 2;
		return true;
	}

	return false;
}

/**
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_GH_To_F()
{
	// the cases covered here would fall under
	// the GH_To_F rule below otherwise
	if(Encode_GH_Special_Cases())
	{
		return true;
	}
	else
	{
		//e.g., 'laugh', 'cough', 'rough', 'tough'
		if((m_current > 2)
			AND (m_inWord[m_current - 1] == 'U')
			AND IsVowel(m_current - 2)
			AND StringAt((m_current - 3), 1, "C", "G", "L", "R", "T", "N", "S", "")
			AND !StringAt((m_current - 4), 8, "BREUGHEL", "FLAUGHER", ""))
		{
			MetaphAdd("F");
			m_current += 2;
			return true;
		}
	}

	return false;
}

/**
 * Encode some contexts where "g" is silent
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_Silent_G()
{
	// e.g. "phlegm", "apothegm", "voigt"
	if((((m_current + 1) == m_last)
		AND (StringAt((m_current - 1), 3, "EGM", "IGM", "AGM", "")
			OR StringAt(m_current, 2, "GT", "")))
		OR (StringAt(0, 5, "HUGES", "") && (m_length == 5)))
	{
		m_current++;
		return true;
	}

	// vietnamese names e.g. "Nguyen" but not "Ng"
	if(StringAt(0, 2, "NG", "") AND (m_current != m_last))
	{
		m_current++;
		return true;
	}

	return false;
}

/**
 * ENcode "-GN-"
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_GN()
{
	if(m_inWord[m_current + 1] == 'N')
	{
		// 'align' 'sign', 'resign' but not 'resignation'
		// also 'impugn', 'impugnable', but not 'repugnant'
		if(((m_current > 1)
			AND ((StringAt((m_current - 1), 1, "I", "U", "E", "")
				OR StringAt((m_current - 3), 9, "LORGNETTE", "")
				OR StringAt((m_current - 2), 9, "LAGNIAPPE", "")
				OR StringAt((m_current - 2), 6, "COGNAC", "")
				OR StringAt((m_current - 3), 7, "CHAGNON", "")
				OR StringAt((m_current - 5), 9, "COMPAGNIE", "")
				OR StringAt((m_current - 4), 6, "BOLOGN", ""))
			// Exceptions: following are cases where 'G' is pronounced
			// in "assign" 'g' is silent, but not in "assignation"
			AND !(StringAt((m_current + 2), 5, "ATION", "")
				OR StringAt((m_current + 2), 4, "ATOR", "")
				OR StringAt((m_current + 2), 3, "ATE", "ITY", "")
			// exception to exceptions, not pronounced:
			OR (StringAt((m_current + 2), 2, "AN", "AC", "IA", "UM", "")
				AND !(StringAt((m_current - 3), 8, "POIGNANT", "")
						OR StringAt((m_current - 2), 6, "COGNAC", "")))
			OR StringAt(0, 7, "SPIGNER", "STEGNER", "")
			OR (StringAt(0, 5, "SIGNE", "") AND (m_length == 5))
			OR StringAt((m_current - 2), 5, "LIGNI", "LIGNO", "REGNA", "DIGNI", "WEGNE",
											"TIGNE", "RIGNE", "REGNE", "TIGNO", "")
			OR StringAt((m_current - 2), 6, "SIGNAL", "SIGNIF", "SIGNAT", "")
			OR StringAt((m_current - 1), 5, "IGNIT", ""))
			AND !StringAt((m_current - 2), 6, "SIGNET", "LIGNEO", "") ))
			//not e.g. 'cagney', 'magna'
			OR (((m_current + 2) == m_last)
				AND StringAt(m_current, 3, "GNE", "GNA", "")
				AND !StringAt((m_current - 2), 5, "SIGNA", "MAGNA", "SIGNE", "")))
		{
			MetaphAddExactApprox("N", "GN", "N", "KN");
		}
		else
		{
			MetaphAddExactApprox("GN", "KN");
		}
		m_current += 2;
		return true;
	}
	return false;
}

/**
 * Encode "-GL-"
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_GL()
{
	//'tagliaro', 'puglia' BUT add K in alternative
	// since americans sometimes do this
	if(StringAt((m_current + 1), 3, "LIA", "LIO", "LIE", "")
		AND IsVowel(m_current - 1))
	{
		MetaphAddExactApprox("L", "GL", "L", "KL");
		m_current += 2;
		return true;
	}

	return false;
}

/**
 * Encode cases where 'G' is at start of word followed
 * by a "front" vowel e.g. 'E', 'I', 'Y'
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_Initial_G_Front_Vowel()
{
	// 'g' followed by vowel at beginning
	if((m_current == 0) AND Front_Vowel(m_current + 1))
	{
		// special case "gila" as in "gila monster"
		if(StringAt((m_current + 1), 3, "ILA", "")
			AND (m_length == 4))
		{
			MetaphAdd("H");
		}
		else if(Initial_G_Soft())
		{
			MetaphAddExactApprox("J", "G", "J", "K");
		}
		else
		{
			// only code alternate 'J' if front vowel
			if((m_inWord[m_current + 1] == 'E') OR (m_inWord[m_current + 1] == 'I'))
			{
				MetaphAddExactApprox("G", "J", "K", "J");
			}
			else
			{
				MetaphAddExactApprox("G", "K");
			}
		}

		AdvanceCounter(2, 1);
		return true;
	}

	return false;
}

/**
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Initial_G_Soft()
{
	if(((StringAt((m_current + 1), 2, "EL", "EM", "EN", "EO", "ER", "ES", "IA", "IN", "IO", "IP", "IU",
									  "YM", "YN", "YP", "YR", "EE", "")
			OR StringAt((m_current + 1), 3, "IRA", "IRO", ""))
		// except for smaller set of cases where => K, e.g. "gerber"
		AND !(StringAt((m_current + 1), 3, "ELD", "ELT", "ERT", "INZ", "ERH", "ITE", "ERD", "ERL", "ERN",
										   "INT", "EES", "EEK", "ELB", "EER", "")
				OR StringAt((m_current + 1), 4, "ERSH", "ERST", "INSB", "INGR", "EROW", "ERKE", "EREN", "")
				OR StringAt((m_current + 1), 5, "ELLER", "ERDIE", "ERBER", "ESUND", "ESNER", "INGKO",
											    "INKGO", "IPPER", "ESELL", "IPSON", "EEZER", "ERSON", "ELMAN", "")
				OR StringAt((m_current + 1), 6, "ESTALT", "ESTAPO", "INGHAM", "ERRITY", "ERRISH", "ESSNER", "ENGLER", "")
				OR StringAt((m_current + 1), 7, "YNAECOL", "YNECOLO", "ENTHNER", "ERAGHTY", "")
				OR StringAt((m_current + 1), 8, "INGERICH", "EOGHEGAN", "")))
		OR(IsVowel(m_current + 1)
			AND (StringAt((m_current + 1), 3, "EE ", "EEW", "")
				OR (StringAt((m_current + 1), 3, "IGI", "IRA", "IBE", "AOL", "IDE", "IGL", "")
													AND !StringAt((m_current + 1), 5, "IDEON", "") )
				OR StringAt((m_current + 1), 4, "ILES", "INGI", "ISEL", "")
				OR (StringAt((m_current + 1), 5, "INGER", "") AND !StringAt((m_current + 1), 8, "INGERICH", ""))
				OR StringAt((m_current + 1), 5, "IBBER", "IBBET", "IBLET", "IBRAN", "IGOLO", "IRARD", "IGANT", "")
				OR StringAt((m_current + 1), 6, "IRAFFE", "EEWHIZ","")
				OR StringAt((m_current + 1), 7, "ILLETTE", "IBRALTA", ""))))
	{
		return true;
	}

	return false;
}

/**
 * Encode "-NGER-"
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_NGER()
{
	if((m_current > 1)
		AND StringAt((m_current - 1), 4, "NGER", ""))
	{
		// default 'G' => J  such as 'ranger', 'stranger', 'manger', 'messenger', 'orangery', 'granger'
		// 'boulanger', 'challenger', 'danger', 'changer', 'harbinger', 'lounger', 'ginger', 'passenger'
		// except for these the following
		if(!(RootOrInflections(m_inWord, "ANGER")
			OR RootOrInflections(m_inWord, "LINGER")
			OR RootOrInflections(m_inWord, "MALINGER")
			OR RootOrInflections(m_inWord, "FINGER")
			OR (StringAt((m_current - 3), 4, "HUNG", "FING", "BUNG", "WING", "RING", "DING", "ZENG",
											 "ZING", "JUNG", "LONG", "PING", "CONG", "MONG", "BANG",
											 "GANG", "HANG", "LANG", "SANG", "SING", "WANG", "ZANG", "")
				// exceptions to above where 'G' => J
				AND !(StringAt((m_current - 6), 7, "BOULANG", "SLESING", "KISSING", "DERRING", "")
					OR StringAt((m_current - 8), 9, "SCHLESING", "")
					OR StringAt((m_current - 5), 6, "SALING", "BELANG", "")
					OR StringAt((m_current - 6), 7, "BARRING", "")
					OR StringAt((m_current - 6), 9, "PHALANGER", "")
					OR StringAt((m_current - 4), 5, "CHANG", "")))
			OR StringAt((m_current - 4), 5, "STING", "YOUNG", "")
			OR StringAt((m_current - 5), 6, "STRONG", "")
			OR StringAt(0, 3, "UNG", "ENG", "ING", "")
			OR StringAt(0, 6, "SENGER", "")
			OR StringAt(m_current, 6, "GERICH", "")
			OR StringAt((m_current - 3), 6, "WENGER", "MUNGER", "SONGER", "KINGER", "")
			OR StringAt((m_current - 4), 7, "FLINGER", "SLINGER", "STANGER", "STENGER", "KLINGER", "CLINGER", "")
			OR StringAt((m_current - 5), 8, "SPRINGER", "SPRENGER", "")
			OR StringAt((m_current - 3), 7, "LINGERF", "")
			OR StringAt((m_current - 2), 7, "ANGERLY", "ANGERBO", "INGERSO", "") ))
		{
			MetaphAddExactApprox("J", "G", "J", "K");
		}
		else
		{
			MetaphAddExactApprox("G", "J", "K", "J");
		}

		AdvanceCounter(2, 1);
		return true;
	}

	return false;
}

/**
 * Encode "-GER-"
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_GER()
{
	if((m_current > 0)
		AND StringAt((m_current + 1), 2, "ER", ""))
	{
		// Exceptions to 'GE' where 'G' => K
		// e.g. "JAGER", "TIGER", "LIGER", "LAGER", "LUGER", "AUGER", "EAGER", "HAGER", "SAGER"
		if((((m_current == 2) AND IsVowel(m_current - 1) AND !IsVowel(m_current - 2)
				AND !(StringAt((m_current - 2), 5, "PAGER", "WAGER", "NIGER", "ROGER", "LEGER", "CAGER", ""))
			OR StringAt((m_current - 2), 5, "AUGER", "EAGER", "INGER", "YAGER", ""))
			OR StringAt((m_current - 3), 6, "SEEGER", "JAEGER", "GEIGER", "KRUGER", "SAUGER", "BURGER",
											"MEAGER", "MARGER", "RIEGER", "YAEGER", "STEGER", "PRAGER", "SWIGER",
											"YERGER", "TORGER", "FERGER", "HILGER", "ZEIGER", "YARGER",
											"COWGER", "CREGER", "KROGER", "KREGER", "GRAGER", "STIGER", "BERGER", "")
			// 'berger' but not 'bergerac'
			OR (StringAt((m_current - 3), 6, "BERGER", "") && ((m_current + 2) == m_last))
			OR StringAt((m_current - 4), 7, "KREIGER", "KRUEGER", "METZGER", "KRIEGER", "KROEGER", "STEIGER",
											"DRAEGER", "BUERGER", "BOERGER", "FIBIGER", "")
			// e.g. 'harshbarger', 'winebarger'
			OR (StringAt((m_current - 3), 6, "BARGER", "") AND (m_current > 4))
			// e.g. 'weisgerber'
			OR (StringAt(m_current, 6, "GERBER", "") AND (m_current > 0))
			OR StringAt((m_current - 5), 8, "SCHWAGER",	"LYBARGER",	"SPRENGER", "GALLAGER", "WILLIGER", "")
			OR StringAt(0, 4, "HARGER", "")
			OR (StringAt(0, 4, "AGER", "EGER", "") && (m_length == 4))
			OR StringAt((m_current - 1), 6, "YGERNE", "")
			OR StringAt((m_current - 6), 9, "SCHWEIGER", ""))
			AND !(StringAt((m_current - 5), 10, "BELLIGEREN", "")
					OR StringAt(0, 7, "MARGERY", "")
					OR StringAt((m_current - 3), 8, "BERGERAC", "")))
		{
			if(SlavoGermanic())
			{
				MetaphAddExactApprox("G", "K");
			}
			else
			{
				MetaphAddExactApprox("G", "J", "K", "J");
			}
		}
		else
		{
			MetaphAddExactApprox("J", "G", "J", "K");
		}

		AdvanceCounter(2, 1);
		return true;
	}

	return false;
}

/**
 * ENcode "-GEL-"
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_GEL()
{
	// more likely to be "-GEL-" => JL
	if(StringAt((m_current + 1), 2, "EL", "")
		AND (m_current > 0))
	{
		// except for
		// "BAGEL", "HEGEL", "HUGEL", "KUGEL", "NAGEL", "VOGEL", "FOGEL", "PAGEL"
		if(((m_length == 5)
				AND IsVowel(m_current - 1)
				AND !IsVowel(m_current - 2)
				AND !StringAt((m_current - 2), 5, "NIGEL", "RIGEL", ""))
			// or the following as combining forms
			OR StringAt((m_current - 2), 5, "ENGEL", "HEGEL", "NAGEL", "VOGEL", "")
			OR StringAt((m_current - 3), 6, "MANGEL", "WEIGEL", "FLUGEL", "RANGEL", "HAUGEN", "RIEGEL", "VOEGEL", "")
			OR StringAt((m_current - 4), 7, "SPEIGEL", "STEIGEL", "WRANGEL", "SPIEGEL", "")
			OR StringAt((m_current - 4), 8, "DANEGELD", ""))
		{
			if(SlavoGermanic())
			{
				MetaphAddExactApprox("G", "K");
			}
			else
			{
				MetaphAddExactApprox("G", "J", "K", "J");
			}
		}
		else
		{
			MetaphAddExactApprox("J", "G", "J", "K");
		}

		AdvanceCounter(2, 1);
		return true;
	}

	return false;
}

/**
 * Encode "-G-" followed by a vowel when non-initial leter
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_Non_Initial_G_Front_Vowel()
{
	// -gy-, gi-, ge-
	if(StringAt((m_current + 1), 1, "E", "I", "Y", ""))
	{
		// '-ge' at end
		// almost always 'j 'sound
		if(StringAt(m_current, 2, "GE", "") AND (m_current == (m_last - 1)))
		{
			if(Hard_GE_At_End())
			{
				if(SlavoGermanic())
				{
					MetaphAddExactApprox("G", "K");
				}
				else
				{
					MetaphAddExactApprox("G", "J", "K", "J");
				}
			}
			else
			{
				MetaphAdd("J");
			}
		}
		else
		{
			if(Internal_Hard_G())
			{
				// don't encode KG or KK if e.g. "mcgill"
				if(!((m_current == 2) && StringAt(0, 2, "MC", ""))
					  OR ((m_current == 3) && StringAt(0, 3, "MAC", "")))
				{
					if(SlavoGermanic())
					{
						MetaphAddExactApprox("G", "K");
					}
					else
					{
						MetaphAddExactApprox("G", "J", "K", "J");
					}
				}
			}
			else
			{
				MetaphAddExactApprox("J", "G", "J", "K");
			}
		}

		AdvanceCounter(2, 1);
		return true;
	}

	return false;
}

/*
 * Detect german names and other words that have
 * a 'hard' 'g' in the context of "-ge" at end
 *
 * @return true if encoding handled in this routine, false if not
 */
bool Metaphone3::Hard_GE_At_End()
{
	if(StringAt(0, 6, "RENEGE", "STONGE", "STANGE", "PRANGE", "KRESGE", "")
		OR StringAt(0, 5, "BYRGE", "BIRGE", "BERGE", "HAUGE", "")
		OR StringAt(0, 4, "HAGE", "")
		OR StringAt(0, 5, "LANGE", "SYNGE", "BENGE", "RUNGE", "HELGE", "")
		OR StringAt(0, 4, "INGE", "LAGE", ""))
	{
		return true;
	}

	return false;
}

/**
 * Exceptions to default encoding to 'J':
 * encode "-G-" to 'G' in "-g<frontvowel>-" words
 * where we are not at "-GE" at the end of the word
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Internal_Hard_G()
{
	// if not "-GE" at end
	if(!(((m_current + 1) == m_last) AND (m_inWord[m_current + 1] == 'E') )
		AND (Internal_Hard_NG()
			OR Internal_Hard_GEN_GIN_GET_GIT()
			OR Internal_Hard_G_Open_Syllable()
			OR Internal_Hard_G_Other()))
	{
		return true;
	}

	return false;
}

/**
 * Detect words where "-ge-" or "-gi-" get a 'hard' 'g'
 * even though this is usually a 'soft' 'g' context
 *
 * @return true if 'hard' 'g' detected
 *
 */
bool Metaphone3::Internal_Hard_G_Other()
{
	if((StringAt(m_current, 4, "GETH", "GEAR", "GEIS", "GIRL", "GIVI", "GIVE", "GIFT",
							   "GIRD", "GIRT", "GILV", "GILD", "GELD", "")
				AND !StringAt((m_current - 3), 6, "GINGIV", "") )
			// "gish" but not "largish"
			OR (StringAt((m_current + 1), 3, "ISH", "") AND (m_current > 0) AND !StringAt(0, 4, "LARG", ""))
			OR (StringAt((m_current - 2), 5, "MAGED", "MEGID", "") AND !((m_current + 2) == m_last))
			OR StringAt(m_current, 3, "GEZ", "")
			OR StringAt(0, 4, "WEGE", "HAGE", "")
			OR (StringAt((m_current - 2), 6, "ONGEST", "UNGEST", "")
				AND ((m_current + 3) == m_last)
				AND !StringAt((m_current - 3), 7, "CONGEST", ""))
				OR StringAt(0, 5, "VOEGE", "BERGE", "HELGE", "")
			OR (StringAt(0, 4, "ENGE", "BOGY", "") AND (m_length == 4))
			OR StringAt(m_current, 6, "GIBBON", "")
			OR StringAt(0, 10, "CORREGIDOR", "")
			OR StringAt(0, 8, "INGEBORG", "")
			OR (StringAt(m_current, 4, "GILL", "")
				AND (((m_current + 3) == m_last) OR ((m_current + 4) == m_last))
				AND !StringAt(0, 8, "STURGILL", "")))
	{
		return true;
	}

	return false;
}

/**
 * Detect words where "-gy-", "-gie-", "-gee-",
 * or "-gio-" get a 'hard' 'g' even though this is
 * usually a 'soft' 'g' context
 *
 * @return true if 'hard' 'g' detected
 *
 */
bool Metaphone3::Internal_Hard_G_Open_Syllable()
{
	if(StringAt((m_current + 1), 3, "EYE", "")
		OR StringAt((m_current - 2), 4, "FOGY", "POGY", "YOGI", "")
		OR StringAt((m_current - 2), 5, "MAGEE", "HAGIO", "")
		OR StringAt((m_current - 1), 4, "RGEY", "OGEY", "")
		OR StringAt((m_current - 3), 5, "HOAGY", "STOGY", "PORGY", "")
		OR StringAt((m_current - 5), 8, "CARNEGIE", "")
		OR (StringAt((m_current - 1), 4, "OGEY", "OGIE", "") AND ((m_current + 2) == m_last)))
	{
		return true;
	}

	return false;
}

/**
 * Detect a number of contexts, mostly german names, that
 * take a 'hard' 'g'.
 *
 * @return true if 'hard' 'g' detected, false if not
 *
 */
bool Metaphone3::Internal_Hard_GEN_GIN_GET_GIT()
{
	if((StringAt((m_current - 3), 6, "FORGET", "TARGET", "MARGIT", "MARGET", "TURGEN",
									 "BERGEN", "MORGEN", "JORGEN", "HAUGEN", "JERGEN",
									 "JURGEN", "LINGEN", "BORGEN", "LANGEN", "KLAGEN", "STIGER", "BERGER", "")
				AND !StringAt(m_current, 7, "GENETIC", "GENESIS", "")
				AND !StringAt((m_current - 4), 8, "PLANGENT", ""))
		OR (StringAt((m_current - 3), 6, "BERGIN", "FEAGIN", "DURGIN", "") AND ((m_current + 2) == m_last))
		OR (StringAt((m_current - 2), 5, "ENGEN", "") AND !StringAt((m_current + 3), 3, "DER", "ETI", "ESI", ""))
		OR StringAt((m_current - 4), 7, "JUERGEN", "")
		OR StringAt(0, 5, "NAGIN", "MAGIN", "HAGIN", "")
		OR (StringAt(0, 5, "ENGIN", "DEGEN", "LAGEN", "MAGEN", "NAGIN", "") AND (m_length == 5))
		OR (StringAt((m_current - 2), 5, "BEGET", "BEGIN", "HAGEN", "FAGIN",
										 "BOGEN", "WIGIN", "NTGEN", "EIGEN",
										 "WEGEN", "WAGEN", "")
				AND !StringAt((m_current - 5), 8, "OSPHAGEN", "")))
	{
		return true;
	}

	return false;
}

/**
 * Detect a number of contexts of '-ng-' that will
 * take a 'hard' 'g' despite being followed by a
 * front vowel.
 *
 * @return true if 'hard' 'g' detected, false if not
 *
 */
bool Metaphone3::Internal_Hard_NG()
{
	if((StringAt((m_current - 3), 4, "DANG", "FANG", "SING", "")
		// exception to exception
				AND !StringAt((m_current - 5), 8, "DISINGEN", "") )
		OR StringAt(0, 5, "INGEB", "ENGEB", "")
		OR (StringAt((m_current - 3), 4, "RING", "WING", "HANG", "LONG", "")
				AND !(StringAt((m_current - 4), 5, "CRING", "FRING", "ORANG", "TWING", "CHANG", "PHANG", "")
						OR StringAt((m_current - 5), 6, "SYRING", "")
						OR StringAt((m_current - 3), 7, "RINGENC", "RINGENT", "LONGITU", "LONGEVI", "")
						// e.g. 'longino', 'mastrangelo'
						OR (StringAt(m_current, 4, "GELO", "GINO", "") AND ((m_current + 3) == m_last)) ))
		OR (StringAt((m_current - 1), 3, "NGY", "")
		// exceptions to exception
				AND !(StringAt((m_current - 3), 5, "RANGY", "MANGY", "MINGY", "")
						OR StringAt((m_current - 4), 6, "SPONGY", "STINGY", ""))))
	{
		return true;
	}

	return false;
}

/**
 * Encode special case where "-GA-" => J
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_GA_To_J()
{
	// 'margary', 'margarine'
	if((StringAt((m_current - 3), 7, "MARGARY", "MARGARI", "")
		// but not in spanish forms such as "margatita"
		AND !StringAt((m_current - 3), 8, "MARGARIT", ""))
		OR StringAt(0, 4, "GAOL", "")
		OR StringAt((m_current - 2), 5, "ALGAE", ""))
	{
		MetaphAddExactApprox("J", "G", "J", "K");
		AdvanceCounter(2, 1);
		return true;
	}

	return false;
}

/**
 * Encode 'H'
 *
 *
 */
void Metaphone3::Encode_H()
{
	if(Encode_Initial_Silent_H()
		OR Encode_Initial_HS()
		OR Encode_Initial_HU_HW()
		OR Encode_Non_Initial_Silent_H())
	{
		return;
	}

	//only keep if first & before vowel or btw. 2 vowels
	if(!Encode_H_Pronounced())
	{
		//also takes care of 'HH'
		m_current++;
	}
}

/**
 * Encode cases where initial 'H' is not pronounced (in American)
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_Initial_Silent_H()
{
	//'hour', 'herb', 'heir', 'honor'
	if(StringAt((m_current + 1), 3, "OUR", "ERB", "EIR", "")
		OR StringAt((m_current + 1), 4, "ONOR", "")
		OR StringAt((m_current + 1), 5, "ONOUR", "ONEST", ""))
	{
		// british pronounce H in this word
		// americans give it 'H' for the name,
		// no 'H' for the plant
		if(StringAt(m_current, 4, "HERB", ""))
		{
			if(m_encodeVowels)
			{
				MetaphAdd("HA", "A");
			}
			else
			{
				MetaphAdd("H", "A");
			}
		}
		else if((m_current == 0) OR m_encodeVowels)
		{
			MetaphAdd("A");
		}

		m_current++;
		// don't encode vowels twice
		m_current = SkipVowels(m_current);
		return true;
	}

	return false;
}

/**
 * Encode "HS-"
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_Initial_HS()
{
	// old chinese pinyin transliteration
	// e.g., 'HSIAO'
	if ((m_current == 0) AND StringAt(0, 2, "HS", ""))
	{
		MetaphAdd("X");
		m_current += 2;
		return true;
	}

	return false;
}

/**
 * Encode cases where "HU-" is pronounced as part of a vowel diphthong
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_Initial_HU_HW()
{
	// spanish spellings and chinese pinyin transliteration
	if (StringAt(0, 3, "HUA", "HUE", "HWA", ""))
	{
		if(!StringAt(m_current, 4, "HUEY", ""))
		{
			MetaphAdd("A");

			if(!m_encodeVowels)
			{
				m_current += 3;
			}
			else
			{
				m_current++;
				// don't encode vowels twice
				while(IsVowel(m_current) OR (m_inWord[m_current] == 'W'))
				{
					m_current++;
				}
			}
			return true;
		}
	}

	return false;
}

/**
 * Encode cases where 'H' is silent between vowels
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_Non_Initial_Silent_H()
{
	//exceptions - 'h' not pronounced
	// "PROHIB" BUT NOT "PROHIBIT"
	if(StringAt((m_current - 2), 5, "NIHIL", "VEHEM", "LOHEN", "NEHEM",
									"MAHON", "MAHAN", "COHEN", "GAHAN", "")
		OR StringAt((m_current - 3), 6, "GRAHAM", "PROHIB", "FRAHER",
										"TOOHEY", "TOUHEY", "")
		OR StringAt((m_current - 3), 5, "TOUHY", "")
		OR StringAt(0, 9, "CHIHUAHUA", ""))
	{
		if(!m_encodeVowels)
		{
			m_current += 2;
		}
		else
		{
			m_current++;
			// don't encode vowels twice
			m_current = SkipVowels(m_current);
		}
		return true;
	}

	return false;
}

/**
 * Encode cases where 'H' is pronounced
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_H_Pronounced()
{
	if((((m_current == 0)
			OR IsVowel(m_current - 1)
			OR ((m_current > 0)
				AND (m_inWord[m_current - 1] == 'W')))
		AND IsVowel(m_current + 1))
		// e.g. 'alWahhab'
		OR ((m_inWord[m_current + 1] == 'H') && IsVowel(m_current + 2)))
	{
		MetaphAdd("H");
		AdvanceCounter(2, 1);
		return true;
	}

	return false;
}

/**
 * Encode 'J'
 *
 */
void Metaphone3::Encode_J()
{
	if(Encode_Spanish_J()
		OR Encode_Spanish_OJ_UJ())
	{
		return;
	}

	Encode_Other_J();
}

/**
 * Encode cases where initial or medial "j" is in a spanish word or name
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_Spanish_J()
{
	//obvious spanish, e.g. "jose", "san jacinto"
	if((StringAt((m_current + 1), 3, "UAN", "ACI", "ALI", "EFE", "ICA", "IME", "OAQ", "UAR", "")
			AND !StringAt(m_current, 8, "JIMERSON", "JIMERSEN", ""))
		OR (StringAt((m_current + 1), 3, "OSE", "") AND ((m_current + 3) == m_last))
		OR StringAt((m_current + 1), 4, "EREZ", "UNTA", "AIME", "AVIE", "AVIA", "")
		OR StringAt((m_current + 1), 6, "IMINEZ", "ARAMIL", "")
		OR (((m_current + 2) == m_last) AND StringAt((m_current - 2), 5, "MEJIA", ""))
		OR StringAt((m_current - 2), 5, "TEJED", "TEJAD", "LUJAN", "FAJAR", "BEJAR", "BOJOR", "CAJIG",
										"DEJAS", "DUJAR", "DUJAN", "MIJAR", "MEJOR", "NAJAR",
										"NOJOS", "RAJED", "RIJAL", "REJON", "TEJAN", "UIJAN", "")
		OR StringAt((m_current - 3), 8, "ALEJANDR", "GUAJARDO", "TRUJILLO", "")
		OR (StringAt((m_current - 2), 5, "RAJAS", "") AND (m_current > 2))
		OR (StringAt((m_current - 2), 5, "MEJIA", "") AND !StringAt((m_current - 2), 6, "MEJIAN", ""))
		OR StringAt((m_current - 1), 5, "OJEDA", "")
		OR StringAt((m_current - 3), 5, "LEIJA", "MINJA", "")
		OR StringAt((m_current - 3), 6, "VIAJES", "GRAJAL", "")
		OR StringAt(m_current, 8, "JAUREGUI", "")
		OR StringAt((m_current - 4), 8, "HINOJOSA", "")
		OR StringAt(0, 4, "SAN ", "")
		OR (((m_current + 1) == m_last)
		AND (m_inWord[m_current + 1] == 'O')
		// exceptions
		AND !(StringAt(0, 4, "TOJO", "")
			OR StringAt(0, 5, "BANJO", "")
			OR StringAt(0, 6, "MARYJO", ""))))
	{
		// americans pronounce "juan" as 'wan'
		// and "marijuana" and "tijuana" also
		// do not get the 'H' as in spanish, so
		// just treat it like a vowel in these cases
		if(!(StringAt(m_current, 4, "JUAN", "") OR StringAt(m_current, 4, "JOAQ", "")))
		{
			MetaphAdd("H");
		}
		else
		{
			if(m_current == 0)
			{
				MetaphAdd("A");
			}
		}
		AdvanceCounter(2, 1);
		return true;
	}

	// Jorge gets 2nd HARHA. also JULIO, JESUS
	if(StringAt((m_current + 1), 4, "ORGE", "ULIO", "ESUS", "")
		AND !StringAt(0, 6, "JORGEN", ""))
	{
		// get both consonants for "jorge"
		if(StringAt((m_current + 1), 4, "ORGE", ""))
		{
			if(m_encodeVowels)
			{
				MetaphAdd("JARJ", "HARHA");
			}
			else
			{
				MetaphAdd("JRJ", "HRH");
			}
			AdvanceCounter(5, 5);
			return true;
		}

		MetaphAdd("J", "H");
		AdvanceCounter(2, 1);
		return true;
	}

	return false;
}

/**
 * Encode cases where 'J' is clearly in a german word or name
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_German_J()
{
	if(StringAt((m_current + 1), 2, "AH", "")
		OR (StringAt((m_current + 1), 5, "OHANN", "") AND ((m_current + 5) == m_last))
		OR (StringAt((m_current + 1), 3, "UNG", "") AND !StringAt((m_current + 1), 4, "UNGL", ""))
		OR StringAt((m_current + 1), 3, "UGO", ""))
	{
		MetaphAdd("A");
		AdvanceCounter(2, 1);
		return true;
	}

	return false;
}

/**
 * Encode "-JOJ-" and "-JUJ-" as spanish words
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_Spanish_OJ_UJ()
{
	if(StringAt((m_current + 1), 5, "OJOBA", "UJUY ", ""))
	{
		if(m_encodeVowels)
		{
			MetaphAdd("HAH");
		}
		else
		{
			MetaphAdd("HH");
		}

		AdvanceCounter(4, 3);
		return true;
	}

	return false;
}

/**
 * Encode 'J' => J
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_J_To_J()
{
	if(IsVowel(m_current + 1))
	{
		if((m_current == 0)
			AND Names_Beginning_With_J_That_Get_Alt_Y())
		{
			// 'Y' is a vowel so encode
			// is as 'A'
			if(m_encodeVowels)
			{
				MetaphAdd("JA", "A");
			}
			else
			{
				MetaphAdd("J", "A");
			}
		}
		else
		{
			if(m_encodeVowels)
			{
				MetaphAdd("JA");
			}
			else
			{
				MetaphAdd("J");
			}
		}

		m_current++;
		m_current = SkipVowels(m_current);
		return false;
	}
	else
	{
		MetaphAdd("J");
		m_current++;
		return true;
	}

	return false;
}

/**
 * Encode 'J' toward end in spanish words
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_Spanish_J_2()
{
	// spanish forms e.g. "brujo", "badajoz"
	if((((m_current - 2) == 0)
		AND StringAt((m_current - 2), 4, "BOJA", "BAJA", "BEJA", "BOJO", "MOJA", "MOJI", "MEJI", ""))
		OR (((m_current - 3) == 0)
		AND StringAt((m_current - 3), 5, "FRIJO", "BRUJO", "BRUJA", "GRAJE", "GRIJA", "LEIJA", "QUIJA", ""))
		OR (((m_current + 3) == m_last)
		AND StringAt((m_current - 1), 5, "AJARA", ""))
		OR (((m_current + 2) == m_last)
		AND StringAt((m_current - 1), 4, "AJOS", "EJOS", "OJAS", "OJOS", "UJON", "AJOZ", "AJAL", "UJAR", "EJON", "EJAN", ""))
		OR (((m_current + 1) == m_last)
		AND (StringAt((m_current - 1), 3, "OJA", "EJA", "") AND !StringAt(0, 4, "DEJA", ""))))
	{
		MetaphAdd("H");
		AdvanceCounter(2, 1);
		return true;
	}

	return false;
}

/**
 * Encode 'J' as vowel in some exception cases
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_J_As_Vowel()
{
	if(StringAt(m_current, 5, "JEWSK", ""))
	{
		MetaphAdd("J", "");
		return true;
	}

	// e.g. "stijl", "sejm" - dutch, scandanavian, and eastern european spellings
	if((StringAt((m_current + 1), 1, "L", "T", "K", "S", "N", "M", "")
			// except words from hindi and arabic
			AND !StringAt((m_current + 2), 1, "A", ""))
		OR StringAt(0, 9, "HALLELUJA", "LJUBLJANA", "")
		OR StringAt(0, 4, "LJUB", "BJOR", "")
		OR StringAt(0, 5, "HAJEK", "")
		OR StringAt(0, 3, "WOJ", "")
		// e.g. 'fjord'
		OR StringAt(0, 2, "FJ", "")
		// e.g. 'rekjavik', 'blagojevic'
		OR StringAt(m_current, 5, "JAVIK", "JEVIC", "")
		OR (((m_current + 1) == m_last) AND StringAt(0, 5, "SONJA", "TANJA", "TONJA", "")))
	{
		return true;
	}
	return false;
}

/**
 * Call routines to encode 'J', in proper order
 *
 */
void Metaphone3::Encode_Other_J()
{
	if(m_current == 0)
	{
		if(Encode_German_J())
		{
			return;
		}
		else
		{
			if(Encode_J_To_J())
			{
				return;
			}
		}
	}
	else
	{
		if(Encode_Spanish_J_2())
		{
			return;
		}
		else if(!Encode_J_As_Vowel())
		{
			MetaphAdd("J");
		}

		//it could happen! e.g. "hajj"
		// eat redundant 'J'
		if(m_inWord[m_current + 1] == 'J')
		{
			m_current += 2;
		}
		else
		{
			m_current++;
		}
	}
}

/**
 * Encode 'K'
 *
 *
 */
void Metaphone3::Encode_K()
{
	if(!Encode_Silent_K())
	{
		MetaphAdd("K");

		// eat redundant 'K's and 'Q's
		if((m_inWord[m_current + 1] == 'K')
			OR (m_inWord[m_current + 1] == 'Q'))
		{
			m_current += 2;
		}
		else
		{
			m_current++;
		}
	}
}

/**
 * Encode cases where 'K' is not pronounced
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_Silent_K()
{
    //skip this except for special cases
    if((m_current == 0)
		AND StringAt(m_current, 2, "KN", ""))
    {
        if(!(StringAt((m_current + 2), 5, "ESSET", "IEVEL", "") OR StringAt((m_current + 2), 3, "ISH", "") ))
        {
            m_current += 1;
			return true;
        }
    }

	// e.g. "know", "knit", "knob"
	if((StringAt((m_current + 1), 3, "NOW", "NIT", "NOT", "NOB", "")
			// exception, "slipknot" => SLPNT but "banknote" => PNKNT
			AND !StringAt(0, 8, "BANKNOTE", ""))
		OR StringAt((m_current + 1), 4, "NOCK", "NUCK", "NIFE", "NACK", "")
		OR StringAt((m_current + 1), 5, "NIGHT", ""))
	{
		// N already encoded before
		// e.g. "penknife"
		if ((m_current > 0) AND m_inWord[m_current - 1] == 'N')
		{
			m_current += 2;
		}
		else
		{
			m_current++;
		}

		return true;
	}

	return false;
}

/**
 * Encode 'L'
 *
 * Includes special vowel transposition
 * encoding, where 'LE' => AL
 *
 */
void Metaphone3::Encode_L()
{
	// logic below needs to know this
	// after 'm_current' variable changed
	int save_current = m_current;

	Interpolate_Vowel_When_Cons_L_At_End();

	if(Encode_LELY_To_L()
		OR Encode_COLONEL()
		OR Encode_French_AULT()
		OR Encode_French_EUIL()
		OR Encode_French_OULX()
		OR Encode_Silent_L_In_LM()
		OR Encode_Silent_L_In_LK_LV()
		OR Encode_Silent_L_In_OULD())
	{
		return;
	}

	if(Encode_LL_As_Vowel_Cases())
	{
		return;
	}

	Encode_LE_Cases(save_current);
}

/**
 * Cases where an L follows D, G, or T at the
 * end have a schwa pronounced before the L
 *
 */
void Metaphone3::Interpolate_Vowel_When_Cons_L_At_End()
{
	if(m_encodeVowels)
	{
		// e.g. "ertl", "vogl"
		if((m_current == m_last)
			&& StringAt((m_current - 1), 1, "D", "G", "T", ""))
		{
			MetaphAdd("A");
		}
	}
}

/**
 * Catch cases where 'L' spelled twice but pronounced
 * once, e.g., 'DOCILELY' => TSL
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_LELY_To_L()
{
	// e.g. "agilely", "docilely"
	if(StringAt((m_current - 1), 5, "ILELY", "")
		AND ((m_current + 3) == m_last))
	{
		MetaphAdd("L");
		m_current += 3;
		return true;
	}

	return false;
}

/**
 * Encode special case "colonel" => KRNL. Can somebody tell
 * me how this pronounciation came to be?
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_COLONEL()
{
	if(StringAt((m_current - 2), 7, "COLONEL", ""))
	{
		MetaphAdd("R");
		m_current += 2;
		return true;
	}

	return false;
}

/**
 * Encode "-AULT-", found in a french names
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_French_AULT()
{
	// e.g. "renault" and "foucault", well known to americans, but not "fault"
	if(((m_current > 3)
		AND (StringAt((m_current - 3), 5, "RAULT", "NAULT", "BAULT", "SAULT", "GAULT", "CAULT", "")
			OR StringAt((m_current - 4), 6, "REAULT", "RIAULT", "NEAULT", "BEAULT", "")))
		AND !(RootOrInflections(m_inWord, "ASSAULT")
			OR StringAt((m_current - 8), 10, "SOMERSAULT","")
			OR StringAt((m_current - 9), 11, "SUMMERSAULT", "")))
	{
		m_current += 2;
		return true;
	}

	return false;
}

/**
 * Encode "-EUIL-", always found in a french word
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_French_EUIL()
{
	// e.g. "auteuil"
	if(StringAt((m_current - 3), 4, "EUIL", "") AND (m_current == m_last))
	{
		m_current++;
		return true;
	}

	return false;
}

/**
 * Encode "-OULX", always found in a french word
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_French_OULX()
{
	// e.g. "proulx"
	if(StringAt((m_current - 2), 4, "OULX", "") && ((m_current + 1) == m_last))
	{
		m_current += 2;
		return true;
	}
	return false;
}

/**
 * Encodes contexts where 'L' is not pronounced in "-LM-"
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_Silent_L_In_LM()
{
	if(StringAt(m_current, 2, "LM", "LN", ""))
	{
		// e.g. "lincoln", "holmes", "psalm", "salmon"
		if((StringAt((m_current - 2), 4, "COLN", "CALM", "BALM", "MALM", "PALM", "")
			OR (StringAt((m_current - 1), 3, "OLM", "") AND ((m_current + 1) == m_last))
			OR StringAt((m_current - 3), 5, "PSALM", "QUALM", "")
			OR StringAt((m_current - 2), 6,  "SALMON", "HOLMES", "")
			OR StringAt((m_current - 1), 6,  "ALMOND", "")
			OR ((m_current == 1) AND StringAt((m_current - 1), 4, "ALMS", "") ))
			AND (!StringAt((m_current + 2), 1, "A", "")
				AND !StringAt((m_current - 2), 5, "BALMO", "")
				AND !StringAt((m_current - 2), 6, "PALMER", "PALMOR", "BALMER", "")
				AND !StringAt((m_current - 3), 5, "THALM", "")))
		{
			m_current++;
			return true;
		}
		else
		{
			MetaphAdd("L");
			m_current++;
			return true;
		}
	}

	return false;
}

/**
 * Encodes contexts where '-L-' is silent in 'LK', 'LV'
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_Silent_L_In_LK_LV()
{
	if((StringAt((m_current - 2), 4, "WALK", "YOLK", "FOLK", "HALF", "TALK", "CALF", "BALK", "CALK", "")
		OR (StringAt((m_current - 2), 4, "POLK", "")
			AND !StringAt((m_current - 2), 5, "POLKA", "WALKO", ""))
		OR (StringAt((m_current - 2), 4, "HALV", "")
			AND !StringAt((m_current - 2), 5, "HALVA", "HALVO", ""))
		OR (StringAt((m_current - 3), 5, "CAULK", "CHALK", "BAULK", "FAULK", "")
			AND !StringAt((m_current - 4), 6, "SCHALK", ""))
		OR (StringAt((m_current - 2), 5, "SALVE", "CALVE", "")
		OR StringAt((m_current - 2), 6, "SOLDER", ""))
		// exceptions to above cases where 'L' is usually pronounced
		AND !StringAt((m_current - 2), 6, "SALVER", "CALVER", ""))
		AND !StringAt((m_current - 5), 9, "GONSALVES", "GONCALVES", "")
		AND !StringAt((m_current - 2), 6, "BALKAN", "TALKAL", "")
		AND !StringAt((m_current - 3), 5, "PAULK", "CHALF", ""))
	{
		m_current++;
		return true;
	}

	return false;
}

/**
 * Encode 'L' in contexts of "-OULD-" where it is silent
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_Silent_L_In_OULD()
{
	//'would', 'could'
	if(StringAt((m_current - 3), 5, "WOULD", "COULD", "")
		OR (StringAt((m_current - 4), 6, "SHOULD", "")
			AND !StringAt((m_current - 4), 8, "SHOULDER", "")))
	{
		MetaphAddExactApprox("D", "T");
		m_current += 2;
		return true;
	}

	return false;
}

/**
 * Encode "-ILLA-" and "-ILLE-" in spanish and french
 * contexts were americans know to pronounce it as a 'Y'
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_LL_As_Vowel_Special_Cases()
{
	if(StringAt((m_current - 5), 8, "TORTILLA", "")
		OR StringAt((m_current - 8), 11, "RATATOUILLE", "")
		// e.g. 'guillermo', "veillard"
		OR (StringAt(0, 5, "GUILL", "VEILL", "GAILL", "")
		    // 'guillotine' usually has '-ll-' pronounced as 'L' in english
			AND !(StringAt((m_current - 3), 7, "GUILLOT", "GUILLOR", "GUILLEN", "")
				 OR (StringAt(0, 5, "GUILL", "") AND (m_length == 5))))
		// e.g. "brouillard", "gremillion"
		OR StringAt(0, 7, "BROUILL", "GREMILL", "ROBILL", "")
		// e.g. 'mireille'
		OR (StringAt((m_current - 2), 5, "EILLE", "")
			AND ((m_current + 2) == m_last)
			// exception "reveille" usually pronounced as 're-vil-lee'
			AND !StringAt((m_current - 5), 8, "REVEILLE", "")))
	{
		m_current += 2;
		return true;
	}

	return false;
}

/**
 * Encode other spanish cases where "-LL-" is pronounced as 'Y'
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_LL_As_Vowel()
{
	//spanish e.g. "cabrillo", "gallegos" but also "gorilla", "ballerina" -
	// give both pronounciations since an american might pronounce "cabrillo"
	// in the spanish or the american fashion.
	if((((m_current + 3) == m_length)
		AND StringAt((m_current - 1), 4, "ILLO", "ILLA", "ALLE", ""))
		OR (((StringAt((m_last - 1), 2, "AS", "OS", "")
			OR StringAt(m_last, 2, "AS", "OS", "")
			OR StringAt(m_last, 1, "A", "O", ""))
				AND StringAt((m_current - 1), 2, "AL", "IL", ""))
			AND !StringAt((m_current - 1), 4, "ALLA", ""))
		OR StringAt(0, 5, "VILLE", "VILLA", "")
		OR StringAt(0, 8, "GALLARDO", "VALLADAR", "MAGALLAN", "CAVALLAR", "BALLASTE", "")
		OR StringAt(0, 3, "LLA", ""))
	{
		MetaphAdd("L", "");
		m_current += 2;
		return true;
	}

	return false;
}

/**
 * Call routines to encode "-LL-", in proper order
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_LL_As_Vowel_Cases()
{
	if(m_inWord[m_current + 1] == 'L')
	{
		if(Encode_LL_As_Vowel_Special_Cases())
		{
			return true;
		}
		else if(Encode_LL_As_Vowel())
		{
			return true;
		}
		m_current += 2;

	}
	else
	{
		m_current++;
	}

	return false;
}

/**
 * Encode vowel-encoding cases where "-LE-" is pronounced "-EL-"
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_Vowel_LE_Transposition(int save_current)
{
	// transposition of vowel sound and L occurs in many words,
	// e.g. "bristle", "dazzle", "goggle" => KAKAL
	if(m_encodeVowels
		AND (save_current > 1)
		AND !IsVowel(save_current - 1)
		AND (m_inWord[save_current + 1] == 'E')
		AND (m_inWord[save_current - 1] != 'L')
		AND (m_inWord[save_current - 1] != 'R')
		// lots of exceptions to this:
		AND !IsVowel(save_current + 2)
		AND !StringAt(0, 7, "ECCLESI", "COMPLEC", "COMPLEJ", "ROBLEDO", "")
		AND !StringAt(0, 5, "MCCLE", "MCLEL", "")
		AND !StringAt(0, 6, "EMBLEM", "KADLEC", "")
		AND !(((save_current + 2) == m_last) AND StringAt(save_current, 3, "LET", ""))
		AND !StringAt(save_current, 7, "LETTING", "")
		AND !StringAt(save_current, 6, "LETELY", "LETTER", "LETION", "LETIAN", "LETING", "LETORY", "")
		AND !StringAt(save_current, 5, "LETUS", "LETIV", "")
		AND !StringAt(save_current, 4, "LESS", "LESQ", "LECT", "LEDG", "LETE", "LETH", "LETS", "LETT", "")
		AND !StringAt(save_current, 3, "LEG", "LER", "LEX", "")
		// e.g. "complement" !=> KAMPALMENT
		AND !(StringAt(save_current, 6, "LEMENT", "")
			AND !(StringAt((m_current - 5), 6, "BATTLE", "TANGLE", "PUZZLE", "RABBLE", "BABBLE", "")
				OR StringAt((m_current - 4), 5, "TABLE", "")))
		AND !(((save_current + 2) == m_last) AND StringAt((save_current - 2), 5, "OCLES", "ACLES", "AKLES", ""))
		AND !StringAt((save_current - 3), 5, "LISLE", "AISLE", "")
		AND !StringAt(0, 4, "ISLE", "")
		AND !StringAt(0, 6, "ROBLES", "")
		AND !StringAt((save_current - 4), 7, "PROBLEM", "RESPLEN", "")
		AND !StringAt((save_current - 3), 6, "REPLEN", "")
		AND !StringAt((save_current - 2), 4, "SPLE", "")
		AND (m_inWord[save_current - 1] != 'H')
		AND (m_inWord[save_current - 1] != 'W'))
	{
		MetaphAdd("AL");
		flag_AL_inversion = true;

		// eat redundant 'L'
		if(m_inWord[save_current + 2] == 'L')
		{
			m_current = save_current + 3;
		}
		return true;
	}

	return false;
}

/**
 * Encode special vowel-encoding cases where 'E' is not
 * silent at the end of a word as is the usual case
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_Vowel_Preserve_Vowel_After_L(int save_current)
{
	// an example of where the vowel would NOT need to be preserved
	// would be, say, "hustled", where there is no vowel pronounced
	// between the 'l' and the 'd'
	if(m_encodeVowels
		AND !IsVowel(save_current - 1)
		AND (m_inWord[save_current + 1] == 'E')
		AND (save_current > 1)
		AND ((save_current + 1) != m_last)
		AND !(StringAt((save_current + 1), 2, "ES", "ED", "")
		AND ((save_current + 2) == m_last))
		AND !StringAt((save_current - 1), 5, "RLEST", "") )
	{
		MetaphAdd("LA");
		m_current = SkipVowels(m_current);
		return true;
	}

	return false;
}

/**
 * Call routines to encode "-LE-", in proper order
 *
 * @param save_current index of actual current letter
 *
 */
void Metaphone3::Encode_LE_Cases(int save_current)
{
	if(Encode_Vowel_LE_Transposition(save_current))
	{
		return;
	}
	else
	{
		if(Encode_Vowel_Preserve_Vowel_After_L(save_current))
		{
			return;
		}
		else
		{
			MetaphAdd("L");
		}
	}
}

/**
 * Encode "-M-"
 *
 */
void Metaphone3::Encode_M()
{
	if(Encode_Silent_M_At_Beginning()
		OR Encode_MR_And_MRS()
		OR Encode_MAC()
		OR Encode_MPT())
	{
		return;
	}

	// Silent 'B' should really be handled
	// under 'B", not here under 'M'!
	Encode_MB();

	MetaphAdd("M");
}

/**
 * Encode cases where 'M' is silent at beginning of word
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_Silent_M_At_Beginning()
{
	//skip these when at start of word
    if((m_current == 0)
		AND StringAt(m_current, 2, "MN", ""))
	{
        m_current += 1;
		return true;
	}

	return false;
}

/**
 * Encode special cases "Mr." and "Mrs."
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_MR_And_MRS()
{
	if((m_current == 0) AND StringAt(m_current, 2, "MR", ""))
	{
		// exceptions for "mr." and "mrs."
		if((m_length == 2) AND StringAt(m_current, 2, "MR", ""))
		{
			if(m_encodeVowels)
			{
				MetaphAdd("MASTAR");
			}
			else
			{
				MetaphAdd("MSTR");
			}
			m_current += 2;
			return true;
		}
		else if((m_length == 3) AND StringAt(m_current, 3, "MRS", ""))
		{
			if(m_encodeVowels)
			{
				MetaphAdd("MASAS");
			}
			else
			{
				MetaphAdd("MSS");
			}
			m_current += 3;
			return true;
		}
	}

	return false;
}

/**
 * Encode "Mac-" and "Mc-"
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_MAC()
{
	// should only find scottish names e.g. 'macintosh'
	if((m_current == 0)
		AND (StringAt(0, 7, "MACIVER", "MACEWEN", "")
		OR StringAt(0, 8, "MACELROY", "MACILROY", "")
		OR StringAt(0, 9, "MACINTOSH", "")
		OR StringAt(0, 2, "MC", "")	))
	{
		if(m_encodeVowels)
		{
			MetaphAdd("MAK");
		}
		else
		{
			MetaphAdd("MK");
		}

		if(StringAt(0, 2, "MC", ""))
		{
			if(StringAt((m_current + 2), 1, "K", "G", "Q", "")
				// watch out for e.g. "McGeorge"
				AND !StringAt((m_current + 2), 4, "GEOR", ""))
			{
				m_current += 3;
			}
			else
			{
				m_current += 2;
			}
		}
		else
		{
			m_current += 3;
		}

		return true;
	}

	return false;
}

/**
 * Encode silent 'M' in context of "-MPT-"
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_MPT()
{
	if(StringAt((m_current - 2), 8, "COMPTROL", "")
		OR StringAt((m_current - 4), 7, "ACCOMPT", ""))

	{
		MetaphAdd("N");
		m_current += 2;
		return true;
	}

	return false;
}

/**
 * Test if 'B' is silent in these contexts
 *
 * @return true if 'B' is silent in this context
 *
 */
bool Metaphone3::Test_Silent_MB_1()
{
	// e.g. "LAMB", "COMB", "LIMB", "DUMB", "BOMB"
	// Handle combining roots first
	if (((m_current == 3)
			AND StringAt((m_current - 3), 5, "THUMB", ""))
		OR ((m_current == 2)
			AND StringAt((m_current - 2), 4, "DUMB", "BOMB", "DAMN", "LAMB", "NUMB", "TOMB", "") ))
	{
		return true;
	}

	return false;
}

/**
 * Test if 'B' is pronounced in this context
 *
 * @return true if 'B' is pronounced in this context
 *
 */
bool Metaphone3::Test_Pronounced_MB()
{
	if (StringAt((m_current - 2), 6, "NUMBER", "")
		OR (StringAt((m_current + 2), 1, "A", "")
			AND !StringAt((m_current - 2), 7, "DUMBASS", ""))
		OR StringAt((m_current + 2), 1, "O", "")
		OR StringAt((m_current - 2), 6, "LAMBEN", "LAMBER", "LAMBET", "TOMBIG", "LAMBRE", ""))
	{
		return true;
	}

	return false;
}

/**
 * Test whether "-B-" is silent in these contexts
 *
 * @return true if 'B' is silent in this context
 *
 */
bool Metaphone3::Test_Silent_MB_2()
{
	// 'M' is the current letter
	if ((m_inWord[m_current + 1] == 'B') AND (m_current > 1)
		AND (((m_current + 1) == m_last)
		// other situations where "-MB-" is at end of root
		// but not at end of word. The tests are for standard
		// noun suffixes.
		// e.g. "climbing" => KLMNK
		OR StringAt((m_current + 2), 3, "ING", "ABL", "")
		OR StringAt((m_current + 2), 4, "LIKE", "")
		OR ((m_inWord[m_current + 2] == 'S') AND ((m_current + 2) == m_last))
		OR StringAt((m_current - 5), 7, "BUNCOMB", "")
		// e.g. "bomber",
		OR (StringAt((m_current + 2), 2, "ED", "ER", "")
		AND ((m_current + 3) == m_last)
		AND (StringAt(0, 5, "CLIMB", "PLUMB", "")
		// e.g. "beachcomber"
		OR !StringAt((m_current - 1), 5, "IMBER", "AMBER", "EMBER", "UMBER", ""))
		// exceptions
		AND !StringAt((m_current - 2), 6, "CUMBER", "SOMBER", "") ) ) )
	{
		return true;
	}

	return false;
}

/**
 * Test if 'B' is pronounced in these "-MB-" contexts
 *
 * @return true if "-B-" is pronounced in these contexts
 *
 */
bool Metaphone3::Test_Pronounced_MB_2()
{
	// e.g. "bombastic", "umbrage", "flamboyant"
	if (StringAt((m_current - 1), 5, "OMBAS", "OMBAD", "UMBRA", "")
		OR StringAt((m_current - 3), 4, "FLAM", "") )
	{
		return true;
	}

	return false;
}

/**
 * Tests for contexts where "-N-" is silent when after "-M-"
 *
 * @return true if "-N-" is silent in these contexts
 *
 */
bool Metaphone3::Test_MN()
{

	if ((m_inWord[m_current + 1] == 'N')
		AND (((m_current + 1) == m_last)
		// or at the end of a word but followed by suffixes
		OR (StringAt((m_current + 2), 3, "ING", "EST", "") AND ((m_current + 4) == m_last))
		OR ((m_inWord[m_current + 2] == 'S') AND ((m_current + 2) == m_last))
		OR (StringAt((m_current + 2), 2, "LY", "ER", "ED", "")
			AND ((m_current + 3) == m_last))
		OR StringAt((m_current - 2), 9, "DAMNEDEST", "")
		OR StringAt((m_current - 5), 9, "GODDAMNIT", "") ))
	{
		return true;
	}

	return false;
}

/**
 * Call routines to encode "-MB-", in proper order
 *
 */
void Metaphone3::Encode_MB()
{
	if(Test_Silent_MB_1())
	{
		if(Test_Pronounced_MB())
		{
			m_current++;
		}
		else
		{
			m_current += 2;
		}
	}
	else if(Test_Silent_MB_2())
	{
		if(Test_Pronounced_MB_2())
		{
			m_current++;
		}
		else
		{
			m_current += 2;
		}
	}
	else if(Test_MN())
	{
		m_current += 2;
	}
	else
	{
		// eat redundant 'M'
		if (m_inWord[m_current + 1] == 'M')
		{
			m_current += 2;
		}
		else
		{
			m_current++;
		}
	}
}

/**
 * Encode "-N-"
 *
 */
void Metaphone3::Encode_N()
{
	if(Encode_NCE())
	{
		return;
	}

	// eat redundant 'N'
	if(m_inWord[m_current + 1] == 'N')
	{
		m_current += 2;
	}
	else
	{
		m_current++;
	}

	if (!StringAt((m_current - 3), 8, "MONSIEUR", "")
		// e.g. "aloneness",
		AND !StringAt((m_current - 3), 6, "NENESS", ""))
	{
		MetaphAdd("N");
	}
}

/**
 * Encode "-NCE-" and "-NSE-"
 * "entrance" is pronounced exactly the same as "entrants"
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_NCE()
{
	//'acceptance', 'accountancy'
	if(StringAt((m_current + 1), 1, "C", "S", "")
		AND StringAt((m_current + 2), 1, "E", "Y", "I", "")
		AND (((m_current + 2) == m_last)
			OR (((m_current + 3) == m_last))
				AND (m_inWord[m_current + 3] == 'S')))
	{
		MetaphAdd("NTS");
		m_current += 2;
		return true;
	}

	return false;
}

/**
 * Encode "-P-"
 *
 */
void Metaphone3::Encode_P()
{
	if(Encode_Silent_P_At_Beginning()
	   OR Encode_PT()
	   OR Encode_PH()
	   OR Encode_PPH()
	   OR Encode_RPS()
	   OR Encode_COUP()
	   OR Encode_PNEUM()
	   OR Encode_PSYCH()
	   OR Encode_PSALM())
	{
		return;
	}

	Encode_PB();

	MetaphAdd("P");
}

/**
 * Encode cases where "-P-" is silent at the start of a word
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_Silent_P_At_Beginning()
{
    //skip these when at start of word
    if((m_current == 0)
		AND StringAt(m_current, 2, "PN", "PF", "PS", "PT", ""))
	{
        m_current += 1;
		return true;
	}

	return false;
}

/**
 * Encode cases where "-P-" is silent before "-T-"
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_PT()
{
	// 'pterodactyl', 'receipt', 'asymptote'
	if((m_inWord[m_current + 1] == 'T'))
	{
		if (((m_current == 0) AND StringAt(m_current, 5, "PTERO", ""))
			OR StringAt((m_current - 5), 7, "RECEIPT", "")
			OR StringAt((m_current - 4), 8, "ASYMPTOT", ""))
		{
			MetaphAdd("T");
			m_current += 2;
			return true;
		}
	}
	return false;
}

/**
 * Encode "-PH-", usually as F, with exceptions for
 * cases where it is silent, or where the 'P' and 'T'
 * are pronounced seperately because they belong to
 * two different words in a combining form
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_PH()
{
	if(m_inWord[m_current + 1] == 'H')
	{
		// 'PH' silent in these contexts
		if (StringAt(m_current, 9, "PHTHALEIN", "")
			OR ((m_current == 0) AND StringAt(m_current, 4, "PHTH", ""))
			OR StringAt((m_current - 3), 10, "APOPHTHEGM", ""))
		{
			MetaphAdd("0");
			m_current += 4;
		}
		// combining forms
		//'sheepherd', 'upheaval', 'cupholder'
		else if((m_current > 0)
			AND (StringAt((m_current + 2), 3, "EAD", "OLE", "ELD", "ILL", "OLD", "EAP", "ERD",
											  "ARD", "ANG", "ORN", "EAV", "ART", "")
				OR StringAt((m_current + 2), 4, "OUSE", "")
				OR (StringAt((m_current + 2), 2, "AM", "") AND !StringAt((m_current -1), 5, "LPHAM", ""))
				OR StringAt((m_current + 2), 5, "AMMER", "AZARD", "UGGER", "")
				OR StringAt((m_current + 2), 6, "OLSTER", ""))
					AND !StringAt((m_current - 3), 5, "LYMPH", "NYMPH", ""))
		{
			MetaphAdd("P");
			AdvanceCounter(3, 2);
		}
		else
		{
			MetaphAdd("F");
			m_current += 2;
		}
		return true;
	}

	return false;
}

/**
 * Encode "-PPH-". I don't know why the greek poet's
 * name is transliterated this way...
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_PPH()
{
	// 'sappho'
	if((m_inWord[m_current + 1] == 'P') AND (m_inWord[m_current + 2] == 'H'))
	{
		MetaphAdd("F");
		m_current += 3;
		return true;
	}

	return false;
}

/**
 * Encode "-CORPS-" where "-PS-" not pronounced
 * since the cognate is here from the french
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_RPS()
{
	//'-corps-', 'corpsman'
	if(StringAt((m_current - 3), 5, "CORPS", "")
		AND !StringAt((m_current - 3), 6, "CORPSE", ""))
	{
		m_current += 2;
		return true;
	}

	return false;
}

/**
 * Encode "-COUP-" where "-P-" is not pronounced
 * since the word is from the french
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_COUP()
{
	//'coup'
	if((m_current == m_last)
		AND StringAt((m_current - 3), 4, "COUP", "")
		AND !StringAt((m_current - 5), 6, "RECOUP", ""))
	{
		m_current++;
		return true;
	}

	return false;
}

/**
 * Encode 'P' in non-initial contexts of "-PNEUM-"
 * where is also silent
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_PNEUM()
{
	//'-pneum-'
	if(StringAt((m_current + 1), 4, "NEUM", ""))
	{
		MetaphAdd("N");
		m_current += 2;
		return true;
	}

	return false;
}

/**
 * Encode special case "-PSYCH-" where two encodings need to be
 * accounted for in one syllable, one for the 'PS' and one for
 * the 'CH'
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_PSYCH()
{
	//'-psych-'
	if(StringAt((m_current + 1), 4, "SYCH", ""))
	{
		if(m_encodeVowels)
		{
			MetaphAdd("SAK");
		}
		else
		{
			MetaphAdd("SK");
		}

		m_current += 5;
		return true;
	}

	return false;
}

/**
 * Encode 'P' in context of "-PSALM-", where it has
 * become silent
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_PSALM()
{
	//'-psalm-'
	if(StringAt((m_current + 1), 4, "SALM", ""))
	{
		// go ahead and encode entire word
		if(m_encodeVowels)
		{
			MetaphAdd("SAM");
		}
		else
		{
			MetaphAdd("SM");
		}

		m_current += 5;
		return true;
	}

	return false;
}

/**
 * Eat redundant 'B' or 'P'
 *
 */
void Metaphone3::Encode_PB()
{
	// e.g. "campbell", "raspberry"
	// eat redundant 'P' or 'B'
	if(StringAt((m_current + 1), 1, "P", "B", ""))
	{
		m_current += 2;
	}
	else
	{
		m_current++;
	}
}

/**
 * Encode "-Q-"
 *
 */
void Metaphone3::Encode_Q()
{
	// current pinyin
	if(StringAt(m_current, 3, "QIN", ""))
	{
		MetaphAdd("X");
		m_current++;
		return;
	}

	// eat redundant 'Q'
	if(m_inWord[m_current + 1] == 'Q')
	{
		m_current += 2;
	}
	else
	{
		m_current++;
	}

	MetaphAdd("K");
}

/**
 * Encode "-R-"
 *
 */
void Metaphone3::Encode_R()
{
	if(Encode_RZ())
	{
		return;
	}

	if(!Test_Silent_R())
	{
		if(!Encode_Vowel_RE_Transposition())
		{
			MetaphAdd("R");
		}
	}

	// eat redundant 'R'; also skip 'S' as well as 'R' in "poitiers"
	if((m_inWord[m_current + 1] == 'R') OR StringAt((m_current - 6), 8, "POITIERS", ""))
	{
		m_current += 2;
	}
	else
	{
		m_current++;
	}
}

/**
 * Encode "-RZ-" according
 * to american and polish pronunciations
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_RZ()
{
	if(StringAt((m_current - 2), 4, "GARZ", "KURZ", "MARZ", "MERZ", "HERZ", "PERZ", "WARZ", "")
		OR StringAt(m_current, 5, "RZANO", "RZOLA", "")
		OR StringAt((m_current - 1), 4, "ARZA", "ARZN", ""))
	{
		return false;
	}

	// 'yastrzemski' usually has 'z' silent in
	// united states, but should get 'X' in poland
	if(StringAt((m_current - 4), 11, "YASTRZEMSKI", ""))
	{
		MetaphAdd("R", "X");
		m_current += 2;
		return true;
	}
	// 'BRZEZINSKI' gets two pronunciations
	// in the united states, neither of which
	// are authentically polish
	if(StringAt((m_current - 1), 10, "BRZEZINSKI", ""))
	{
		MetaphAdd("RS", "RJ");
		// skip over 2nd 'Z'
		m_current += 4;
		return true;
	}
	// 'z' in 'rz after voiceless consonant gets 'X'
	// in alternate polish style pronunciation
	else if(StringAt((m_current - 1), 3, "TRZ", "PRZ", "KRZ", "")
			OR (StringAt(m_current, 2, "RZ", "")
				AND (IsVowel(m_current - 1) OR (m_current == 0))))
	{
		MetaphAdd("RS", "X");
		m_current += 2;
		return true;
	}
	// 'z' in 'rz after voiceled consonant, vowel, or at
	// beginning gets 'J' in alternate polish style pronunciation
	else if(StringAt((m_current - 1), 3, "BRZ", "DRZ", "GRZ", ""))
	{
		MetaphAdd("RS", "J");
		m_current += 2;
		return true;
	}
		return false;
}

/**
 * Test whether 'R' is silent in this context
 *
 * @return true if 'R' is silent in this context
 *
 */
bool Metaphone3::Test_Silent_R()
{
	// test cases where 'R' is silent, either because the
	// word is from the french or because it is no longer pronounced.
	// e.g. "rogier", "monsieur", "surburban"
	if(((m_current == m_last)
		// reliably french word ending
		AND StringAt((m_current - 2), 3, "IER", "")
		// e.g. "metier"
		AND (StringAt((m_current - 5), 3, "MET", "VIV", "LUC", "")
		// e.g. "cartier", "bustier"
		OR StringAt((m_current - 6), 4, "CART", "DOSS", "FOUR", "OLIV", "BUST", "DAUM", "ATEL",
										"SONN", "CORM", "MERC", "PELT", "POIR", "BERN", "FORT", "GREN",
										"SAUC", "GAGN", "GAUT", "GRAN", "FORC", "MESS", "LUSS", "MEUN",
										"POTH", "HOLL", "CHEN", "")
		// e.g. "croupier"
		OR StringAt((m_current - 7), 5, "CROUP", "TORCH", "CLOUT", "FOURN", "GAUTH", "TROTT",
										"DEROS", "CHART", "")
		// e.g. "chevalier"
		OR StringAt((m_current - 8), 6, "CHEVAL", "LAVOIS", "PELLET", "SOMMEL", "TREPAN", "LETELL", "COLOMB", "")
		OR StringAt((m_current - 9), 7, "CHARCUT", "")
		OR StringAt((m_current - 10), 8, "CHARPENT", "")))
		OR StringAt((m_current - 2), 7, "SURBURB", "WORSTED", "")
		OR StringAt((m_current - 2), 9, "WORCESTER", "")
		OR StringAt((m_current - 7), 8, "MONSIEUR", "")
		OR StringAt((m_current - 6), 8, "POITIERS", "") )
	{
		return true;
	}

	return false;
}

/**
 * Encode '-re-" as 'AR' in contexts
 * where this is the correct pronunciation
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_Vowel_RE_Transposition()
{
	// -re inversion is just like
	// -le inversion
	// e.g. "fibre" => FABAR or "centre" => SANTAR
	if((m_encodeVowels)
		AND (m_inWord[m_current + 1] == 'E')
		AND (m_length > 3)
		AND !StringAt(0, 5, "OUTRE", "LIBRE", "ANDRE", "")
		AND !(StringAt(0, 4, "FRED", "TRES", "") AND (m_length == 4))
		AND !StringAt((m_current - 2), 5, "LDRED", "LFRED", "NDRED", "NFRED", "NDRES", "TRES", "IFRED", "")
		AND !IsVowel(m_current - 1)
		AND (((m_current + 1) == m_last)
			 OR (((m_current + 2) == m_last)
				AND StringAt((m_current + 2), 1, "D", "S", ""))))
	{
		MetaphAdd("AR");
		return true;
	}

	return false;
}

/**
 * Encode "-S-"
 *
 */
void Metaphone3::Encode_S()
{
	if(Encode_SKJ()
		OR Encode_Special_SW()
		OR Encode_SJ()
		OR Encode_Silent_French_S_Final()
		OR Encode_Silent_French_S_Internal()
		OR Encode_ISL()
		OR Encode_STL()
		OR Encode_Christmas()
		OR Encode_STHM()
		OR Encode_ISTEN()
		OR Encode_Sugar()
		OR Encode_SH()
		OR Encode_SCH()
		OR Encode_SUR()
		OR Encode_SU()
		OR Encode_SSIO()
		OR Encode_SS()
		OR Encode_SIA()
		OR Encode_SIO()
		OR Encode_Anglicisations()
		OR Encode_SC()
		OR Encode_SEA_SUI_SIER()
		OR Encode_SEA())
	{
		return;
	}

	MetaphAdd("S");

	if(StringAt((m_current + 1), 1, "S", "Z", "")
		AND !StringAt((m_current + 1), 2, "SH", ""))
	{
		m_current += 2;
	}
	else
	{
		m_current++;
	}
}

/**
 * Encode a couple of contexts where scandinavian, slavic
 * or german names should get an alternate, native
 * pronunciation of 'SV' or 'XV'
 *
 * @return true if handled
 *
 */
bool Metaphone3::Encode_Special_SW()
{
	if(m_current == 0)
	{
		//
		if(Names_Beginning_With_SW_That_Get_Alt_SV())
		{
			MetaphAdd("S", "SV");
			m_current += 2;
			return true;
		}

		//
		if(Names_Beginning_With_SW_That_Get_Alt_XV())
		{
			MetaphAdd("S", "XV");
			m_current += 2;
			return true;
		}
	}

	return false;
}

/**
 * Encode "-SKJ-" as X ("sh"), since americans pronounce
 * the name Dag Hammerskjold as "hammer-shold"
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_SKJ()
{
	// scandinavian
	if(StringAt(m_current, 4, "SKJO", "SKJU", "")
		AND IsVowel(m_current + 3))
	{
		MetaphAdd("X");
		m_current += 3;
		return true;
	}

	return false;
}

/**
 * Encode initial swedish "SJ-" as X ("sh")
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_SJ()
{
	if(StringAt(0, 2, "SJ", ""))
	{
		MetaphAdd("X");
		m_current += 2;
		return true;
	}

	return false;
}

/**
 * Encode final 'S' in words from the french, where they
 * are not pronounced
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_Silent_French_S_Final()
{
	// "louis" is an exception because it gets two pronuncuations
	if(StringAt(0, 5, "LOUIS", "") && (m_current == m_last))
	{
		MetaphAdd("S", "");
		m_current++;
		return true;
	}

	// french words familiar to americans where final s is silent
	if((m_current == m_last)
		AND (StringAt(0, 4, "YVES", "")
		OR (StringAt(0, 4, "HORS", "") AND (m_current == 3))
		OR StringAt((m_current - 4), 5, "CAMUS", "YPRES", "")
		OR StringAt((m_current - 5), 6, "MESNES", "DEBRIS", "BLANCS", "INGRES", "CANNES", "")
		OR StringAt((m_current - 6), 7, "CHABLIS", "APROPOS", "JACQUES", "ELYSEES", "OEUVRES",
										"GEORGES", "DESPRES", "")
		OR StringAt(0, 8, "ARKANSAS", "FRANCAIS", "CRUDITES", "BRUYERES", "")
		OR StringAt(0, 9, "DESCARTES", "DESCHUTES", "DESCHAMPS", "DESROCHES", "DESCHENES", "")
		OR StringAt(0, 10, "RENDEZVOUS", "")
		OR StringAt(0, 11, "CONTRETEMPS", "DESLAURIERS", ""))
		OR ((m_current == m_last)
				AND StringAt((m_current - 2), 2, "AI", "OI", "UI", "")
				AND !StringAt(0, 4, "LOIS", "LUIS", "")))
	{
		m_current++;
		return true;
	}

	return false;
}

/**
 * Encode non-final 'S' in words from the french where they
 * are not pronounced.
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_Silent_French_S_Internal()
{
	// french words familiar to americans where internal s is silent
	if(StringAt((m_current - 2), 9, "DESCARTES", "")
		OR StringAt((m_current - 2), 7, "DESCHAM", "DESPRES", "DESROCH", "DESROSI", "DESJARD", "DESMARA",
										"DESCHEN", "DESHOTE", "DESLAUR", "")
		OR StringAt((m_current - 2), 6, "MESNES", "")
		OR StringAt((m_current - 5), 8, "DUQUESNE", "DUCHESNE", "")
		OR StringAt((m_current - 7), 10, "BEAUCHESNE", "")
		OR StringAt((m_current - 3), 7, "FRESNEL", "")
		OR StringAt((m_current - 3), 9, "GROSVENOR", "")
		OR StringAt((m_current - 4), 10, "LOUISVILLE", "")
		OR StringAt((m_current - 7), 10, "ILLINOISAN", ""))
	{
		m_current++;
		return true;
	}

	return false;
}

/**
 * Encode silent 'S' in context of "-ISL-"
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_ISL()
{
	//special cases 'island', 'isle', 'carlisle', 'carlysle'
	if((StringAt((m_current - 2), 4, "LISL", "LYSL", "AISL", "")
			AND !StringAt((m_current - 3), 7, "PAISLEY", "BAISLEY", "ALISLAM", "ALISLAH", "ALISLAA", ""))
		OR ((m_current == 1)
			AND ((StringAt((m_current - 1), 4, "ISLE", "")
					OR StringAt((m_current - 1), 5, "ISLAN", ""))
					AND !StringAt((m_current - 1), 5, "ISLEY", "ISLER", ""))))
	{
		m_current++;
		return true;
	}

	return false;
}

/**
 * Encode "-STL-" in contexts where the 'T' is silent. Also
 * encode "-USCLE-" in contexts where the 'C' is silent
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_STL()
{
	//'hustle', 'bustle', 'whistle'
	if((StringAt(m_current, 4, "STLE", "STLI", "")
			AND !StringAt((m_current + 2), 4, "LESS", "LIKE", "LINE", ""))
		OR StringAt((m_current - 3), 7, "THISTLY", "BRISTLY",  "GRISTLY", "")
		// e.g. "corpuscle"
		OR StringAt((m_current - 1), 5, "USCLE", ""))
	{
		// KRISTEN, KRYSTLE, CRYSTLE, KRISTLE all pronounce the 't'
		// also, exceptions where "-LING" is a nominalizing suffix
		if(StringAt(0, 7, "KRISTEN", "KRYSTLE", "CRYSTLE", "KRISTLE", "")
			OR StringAt(0, 11, "CHRISTENSEN", "CHRISTENSON", "")
			OR StringAt((m_current - 3), 9, "FIRSTLING", "")
			OR StringAt((m_current - 2), 8,  "NESTLING",  "WESTLING", ""))
		{
			MetaphAdd("ST");
			m_current += 2;
		}
		else
		{
			if(m_encodeVowels
				AND (m_inWord[m_current + 3] == 'E')
				AND (m_inWord[m_current + 4] != 'R')
				AND !StringAt((m_current + 3), 4, "ETTE", "ETTA", "")
				AND !StringAt((m_current + 3), 2, "EY", ""))
			{
				MetaphAdd("SAL");
				flag_AL_inversion = true;
			}
			else
			{
				MetaphAdd("SL");
			}
			m_current += 3;
		}
		return true;
	}

	return false;
}

/**
 * Encode "christmas". Americans always pronounce this as "krissmuss"
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_Christmas()
{
	//'christmas'
	if(StringAt((m_current - 4), 8, "CHRISTMA", ""))
	{
		MetaphAdd("SM");
		m_current += 3;
		return true;
	}

	return false;
}

/**
 * Encode "-STHM-" in contexts where the 'TH'
 * is silent.
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_STHM()
{
	//'asthma', 'isthmus'
	if(StringAt(m_current, 4, "STHM", ""))
	{
		MetaphAdd("SM");
		m_current += 4;
		return true;
	}

	return false;
}

/**
 * Encode "-ISTEN-" and "-STNT-" in contexts
 * where the 'T' is silent
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_ISTEN()
{
	// 't' is silent in verb, pronounced in name
	if(StringAt(0, 8, "CHRISTEN", ""))
	{
		// the word itself
		if(RootOrInflections(m_inWord, "CHRISTEN")
			OR StringAt(0, 11, "CHRISTENDOM", ""))
		{
			MetaphAdd("S", "ST");
		}
		else
		{
			// e.g. 'christenson', 'christene'
			MetaphAdd("ST");
		}
		m_current += 2;
		return true;
	}

	//e.g. 'glisten', 'listen'
	if(StringAt((m_current - 2), 6, "LISTEN", "RISTEN", "HASTEN", "FASTEN", "MUSTNT", "")
		OR StringAt((m_current - 3), 7,  "MOISTEN", ""))
	{
		MetaphAdd("S");
		m_current += 2;
		return true;
	}

	return false;
}

/**
 * Encode special case "sugar"
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_Sugar()
{
	//special case 'sugar-'
	if(StringAt(m_current, 5, "SUGAR", ""))
	{
		MetaphAdd("X");
		m_current++;
		return true;
	}

	return false;
}

/**
 * Encode "-SH-" as X ("sh"), except in cases
 * where the 'S' and 'H' belong to different combining
 * roots and are therefore pronounced seperately
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_SH()
{
	if(StringAt(m_current, 2, "SH", ""))
	{
		// exception
		if(StringAt((m_current - 2), 8, "CASHMERE", ""))
		{
			MetaphAdd("J");
			m_current += 2;
			return true;
		}

		//combining forms, e.g. 'clotheshorse', 'woodshole'
		if((m_current > 0)
			// e.g. "mishap"
			AND ((StringAt((m_current + 1), 3, "HAP", "") AND ((m_current + 3) == m_last))
			// e.g. "hartsheim", "clothshorse"
			OR StringAt((m_current + 1), 4, "HEIM", "HOEK", "HOLM", "HOLZ", "HOOD", "HEAD", "HEID",
									 "HAAR", "HORS", "HOLE", "HUND", "HELM", "HAWK", "HILL", "")
			// e.g. "dishonor"
			OR StringAt((m_current + 1), 5, "HEART", "HATCH", "HOUSE", "HOUND", "HONOR", "")
			// e.g. "mishear"
			OR (StringAt((m_current + 2), 3, "EAR", "") AND ((m_current + 4) == m_last))
			// e.g. "hartshorn"
			OR (StringAt((m_current + 2), 3, "ORN", "") AND !StringAt((m_current - 2), 7, "UNSHORN", ""))
			// e.g. "newshour" but not "bashour", "manshour"
			OR (StringAt((m_current + 1), 4, "HOUR", "")
				AND !(StringAt(0, 7, "BASHOUR", "") OR StringAt(0, 8, "MANSHOUR", "") OR StringAt(0, 6, "ASHOUR", "") ))
			// e.g. "dishonest", "grasshopper"
			OR StringAt((m_current + 2), 5, "ARMON", "ONEST", "ALLOW", "OLDER", "OPPER", "EIMER", "ANDLE", "ONOUR", "")
			// e.g. "dishabille", "transhumance"
			OR StringAt((m_current + 2), 6, "ABILLE", "UMANCE", "ABITUA", "")))
		{
			if (!StringAt((m_current - 1), 1, "S", ""))
				MetaphAdd("S");
		}
		else
		{
			MetaphAdd("X");
		}

		m_current += 2;
		return true;
	}

	return false;
}

/**
 * Encode "-SCH-" in cases where the 'S' is pronounced
 * seperately from the "CH", in words from the dutch, italian,
 * and greek where it can be pronounced SK, and german words
 * where it is pronounced X ("sh")
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_SCH()
{
	// these words were combining forms many centuries ago
	if(StringAt((m_current + 1), 2, "CH", ""))
	{
		if((m_current > 0)
			// e.g. "mischief", "escheat"
			AND (StringAt((m_current + 3), 3, "IEF", "EAT", "")
			// e.g. "mischance"
			OR StringAt((m_current + 3), 4, "ANCE", "ARGE", "")
			// e.g. "eschew"
			OR StringAt(0, 6, "ESCHEW", "")))
		{
			MetaphAdd("S");
			m_current++;
			return true;
		}

		//Schlesinger's rule
		//dutch, danish, italian, greek origin, e.g. "school", "schooner", "schiavone", "schiz-"
		if((StringAt((m_current + 3), 2, "OO", "ER", "EN", "UY", "ED", "EM", "IA", "IZ", "IS", "OL", "")
				AND !StringAt(m_current, 6, "SCHOLT", "SCHISL", "SCHERR", ""))
			OR StringAt((m_current + 3), 3, "ISZ", "")
			OR (StringAt((m_current -1), 6, "ESCHAT", "ASCHIN", "ASCHAL", "ISCHAE", "ISCHIA", "")
				AND !StringAt((m_current - 2), 8, "FASCHING", ""))
			OR (StringAt((m_current - 1), 5, "ESCHI", "")  && ((m_current + 3) == m_last))
			OR (m_inWord[m_current + 3] == 'Y'))
		{
			// e.g. "schermerhorn", "schenker", "schistose"
			if(StringAt((m_current + 3), 2, "ER", "EN", "IS", "")
				AND (((m_current + 4) == m_last)
					OR StringAt((m_current + 3), 3, "ENK", "ENB", "IST", "")))
			{
				MetaphAdd("X", "SK");
			}
			else
			{
				MetaphAdd("SK");
			}
			m_current += 3;
			return true;
		}
		else
		{
			MetaphAdd("X");
			m_current += 3;
			return true;
		}
	}

	return false;
}

/**
 * Encode "-SUR<E,A,Y>-" to J, unless it is at the beginning,
 * or preceeded by 'N', 'K', or "NO"
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_SUR()
{
	// 'erasure', 'usury'
	if(StringAt((m_current + 1), 3, "URE", "URA", "URY", ""))
	{
		//'sure', 'ensure'
		if ((m_current == 0)
			OR StringAt((m_current - 1), 1, "N", "K", "")
			OR StringAt((m_current - 2), 2, "NO", ""))
		{
			MetaphAdd("X");
		}
		else
		{
			MetaphAdd("J");
		}

		AdvanceCounter(2, 1);
		return true;
	}

	return false;
}

/**
 * Encode "-SU<O,A>-" to X ("sh") unless it is preceeded by
 * an 'R', in which case it is encoded to S, or it is
 * preceeded by a vowel, in which case it is encoded to J
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_SU()
{
	//'sensuous', 'consensual'
	if(StringAt((m_current + 1), 2, "UO", "UA", "") AND (m_current != 0))
	{
		// exceptions e.g. "persuade"
		if(StringAt((m_current - 1), 4, "RSUA", ""))
		{
			MetaphAdd("S");
		}
		// exceptions e.g. "casual"
		else if(IsVowel(m_current - 1))
		{
			MetaphAdd("J", "S");
		}
		else
		{
			MetaphAdd("X", "S");
		}

		AdvanceCounter(3, 1);
		return true;
	}

	return false;
}

/**
 * Encodes "-SSIO-" in contexts where it is pronounced
 * either J or X ("sh")
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_SSIO()
{
	if(StringAt((m_current + 1), 4, "SION", ""))
	{
		//"abcission"
		if (StringAt((m_current - 2), 2, "CI", ""))
		{
			MetaphAdd("J");
		}
		//'mission'
		else
		{
			if (IsVowel(m_current - 1))
			{
				MetaphAdd("X");
			}
		}

		AdvanceCounter(4, 2);
		return true;
	}

	return false;
}

/**
 * Encode "-SS-" in contexts where it is pronounced X ("sh")
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_SS()
{
	// e.g. "russian", "pressure"
	if(StringAt((m_current - 1), 5, "USSIA", "ESSUR", "ISSUR", "ISSUE", "")
		// e.g. "hessian", "assurance"
		OR StringAt((m_current - 1), 6, "ESSIAN", "ASSURE", "ASSURA", "ISSUAB", "ISSUAN", "ASSIUS", ""))
	{
		MetaphAdd("X");
		AdvanceCounter(3, 2);
		return true;
	}

	return false;
}

/**
 * Encodes "-SIA-" in contexts where it is pronounced
 * as X ("sh"), J, or S
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_SIA()
{
	// e.g. "controversial", also "fuchsia", "ch" is silent
	if(StringAt((m_current - 2), 5, "CHSIA", "")
		OR StringAt((m_current - 1), 5, "RSIAL", ""))
	{
		MetaphAdd("X");
		AdvanceCounter(3, 1);
		return true;
	}

	// names generally get 'X' where terms, e.g. "aphasia" get 'J'
	if((StringAt(0, 6, "ALESIA", "ALYSIA", "ALISIA", "STASIA", "")
			AND (m_current == 3)
			AND !StringAt(0, 9, "ANASTASIA", ""))
		OR StringAt((m_current - 5), 9, "DIONYSIAN", "")
		OR StringAt((m_current - 5), 8, "THERESIA", ""))
	{
		MetaphAdd("X", "S");
		AdvanceCounter(3, 1);
		return true;
	}

	if((StringAt(m_current, 3, "SIA", "") AND ((m_current + 2) == m_last))
		OR (StringAt(m_current, 4, "SIAN", "") AND ((m_current + 3) == m_last))
		OR StringAt((m_current - 5), 9, "AMBROSIAL", ""))
	{
		if ((IsVowel(m_current - 1) OR StringAt((m_current - 1), 1, "R", ""))
			// exclude compounds based on names, or french or greek words
			AND !(StringAt(0, 5, "JAMES", "NICOS", "PEGAS", "PEPYS", "")
				OR StringAt(0, 6, "HOBBES", "HOLMES", "JAQUES", "KEYNES", "")
				OR StringAt(0, 7, "MALTHUS", "HOMOOUS", "")
				OR StringAt(0, 8, "MAGLEMOS", "HOMOIOUS", "")
				OR StringAt(0, 9, "LEVALLOIS", "TARDENOIS", "")
				OR StringAt((m_current - 4), 5, "ALGES", "") ))
		{
			MetaphAdd("J");
		}
		else
		{
			MetaphAdd("S");
		}

		AdvanceCounter(2, 1);
		return true;
	}
	return false;
}

/**
 * Encodes "-SIO-" in contexts where it is pronounced
 * as J or X ("sh")
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_SIO()
{
	// special case, irish name
	if(StringAt(0, 7, "SIOBHAN", ""))
	{
		MetaphAdd("X");
		AdvanceCounter(3, 1);
		return true;
	}

	if(StringAt((m_current + 1), 3, "ION", ""))
	{
		// e.g. "vision", "version"
		if (IsVowel(m_current - 1) OR StringAt((m_current - 2), 2, "ER", "UR", ""))
		{
			MetaphAdd("J");
		}
		else // e.g. "declension"
		{
			MetaphAdd("X");
		}

		AdvanceCounter(3, 1);
		return true;
	}

	return false;
}

/**
 * Encode cases where "-S-" might well be from a german name
 * and add encoding of german pronounciation in alternate m_metaph
 * so that it can be found in a genealogical search
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_Anglicisations()
{
	//german & anglicisations, e.g. 'smith' match 'schmidt', 'snider' match 'schneider'
	//also, -sz- in slavic language altho in hungarian it is pronounced 's'
	if(((m_current == 0)
		AND StringAt((m_current + 1), 1, "M", "N", "L", ""))
		OR StringAt((m_current + 1), 1, "Z", ""))
	{
		MetaphAdd("S", "X");

		// eat redundant 'Z'
		if(StringAt((m_current + 1), 1, "Z", ""))
		{
			m_current += 2;
		}
		else
		{
			m_current++;
		}

		return true;
	}

	return false;
}

/**
 * Encode "-SC<vowel>-" in contexts where it is silent,
 * or pronounced as X ("sh"), S, or SK
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_SC()
{
	if(StringAt(m_current, 2, "SC", ""))
	{
		// exception 'viscount'
		if(StringAt((m_current - 2), 8, "VISCOUNT", ""))
		{
			m_current += 1;
			return true;
		}

		// encode "-SC<front vowel>-"
		if(StringAt((m_current + 2), 1, "I", "E", "Y", ""))
		{
			// e.g. "conscious"
			if(StringAt((m_current + 2), 4, "IOUS", "")
				// e.g. "prosciutto"
				OR StringAt((m_current + 2), 3, "IUT", "")
				OR StringAt((m_current - 4), 9, "OMNISCIEN", "")
				// e.g. "conscious"
				OR StringAt((m_current - 3), 8, "CONSCIEN", "CRESCEND", "CONSCION", "")
				OR StringAt((m_current - 2), 6, "FASCIS", ""))
			{
				MetaphAdd("X");
			}
			else if(StringAt(m_current, 7, "SCEPTIC", "SCEPSIS", "")
					OR StringAt(m_current, 5, "SCIVV", "SCIRO", "")
					// commonly pronounced this way in u.s.
					OR StringAt(m_current, 6, "SCIPIO", "")
					OR StringAt((m_current - 2), 10, "PISCITELLI", ""))
			{
				MetaphAdd("SK");
			}
			else
			{
				MetaphAdd("S");
			}
			m_current += 2;
			return true;
		}

		MetaphAdd("SK");
		m_current += 2;
		return true;
	}

	return false;
}

/**
 * Encode "-S<EA,UI,IER>-" in contexts where it is pronounced
 * as J
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_SEA_SUI_SIER()
{
	// "nausea" by itself has => NJ as a more likely encoding. Other forms
	// using "nause-" (see Encode_SEA()) have X or S as more familiar pronounciations
	if((StringAt((m_current - 3), 6, "NAUSEA", "") AND ((m_current + 2) == m_last))
		// e.g. "casuistry", "frasier", "hoosier"
		OR StringAt((m_current - 2), 5, "CASUI", "")
		OR (StringAt((m_current - 1), 5, "OSIER", "ASIER", "")
			AND !(StringAt(0, 6, "EASIER","")
				  OR StringAt(0, 5, "OSIER","")
				  OR StringAt((m_current - 2), 6, "ROSIER", "MOSIER", ""))))
	{
		MetaphAdd("J", "X");
		AdvanceCounter(3, 1);
		return true;
	}

	return false;
}

/**
 * Encode cases where "-SE-" is pronounced as X ("sh")
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_SEA()
{
	if((StringAt(0, 4, "SEAN", "") AND ((m_current + 3) == m_last))
		OR (StringAt((m_current - 3), 6, "NAUSEO", "")
		AND !StringAt((m_current - 3), 7, "NAUSEAT", "")))
	{
		MetaphAdd("X");
		AdvanceCounter(3, 1);
		return true;
	}

	return false;
}

/**
 * Encode "-T-"
 *
 */
void Metaphone3::Encode_T()
{
	if(Encode_T_Initial()
		OR Encode_TCH()
		OR Encode_Silent_French_T()
		OR Encode_TUN_TUL_TUA_TUO()
		OR Encode_TUE_TEU_TEOU_TUL_TIE()
		OR Encode_TUR_TIU_Suffixes()
		OR Encode_TI()
		OR Encode_TIENT()
		OR Encode_TSCH()
		OR Encode_TZSCH()
		OR Encode_TH_Pronounced_Separately()
		OR Encode_TTH()
		OR Encode_TH())
	{
		return;
	}

	// eat redundant 'T' or 'D'
	if(StringAt((m_current + 1), 1, "T", "D", ""))
	{
		m_current += 2;
	}
	else
	{
		m_current++;
	}

	MetaphAdd("T");
}

/**
 * Encode some exceptions for initial 'T'
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_T_Initial()
{
	if(m_current == 0)
	{
		// americans usually pronounce "tzar" as "zar"
		if (StringAt((m_current + 1), 3, "SAR", "ZAR", ""))
		{
			m_current++;
			return true;
		}

		// old 'École française d'Extrême-Orient' chinese pinyin where 'ts-' => 'X'
		if (((m_length == 3) && StringAt((m_current + 1), 2, "SO", "SA", "SU", ""))
			OR ((m_length == 4) && StringAt((m_current + 1), 3, "SAO", "SAI", ""))
			OR ((m_length == 5) && StringAt((m_current + 1), 4, "SING", "SANG", "")))
		{
			MetaphAdd("X");
			AdvanceCounter(3, 2);
			return true;
		}

		// "TS<vowel>-" at start can be pronounced both with and without 'T'
		if (StringAt((m_current + 1), 1, "S", "") AND IsVowel(m_current + 2))
		{
			MetaphAdd("TS", "S");
			AdvanceCounter(3, 2);
			return true;
		}

		// e.g. "Tjaarda"
		if (m_inWord[m_current + 1] == 'J')
		{
			MetaphAdd("X");
			AdvanceCounter(3, 2);
			return true;
		}

		// cases where initial "TH-" is pronounced as T and not 0 ("th")
		if ((StringAt((m_current + 1), 2, "HU", "") && (m_length == 3))
			OR StringAt((m_current + 1), 3, "HAI", "HUY", "HAO", "")
			OR StringAt((m_current + 1), 4, "HYME", "HYMY", "HANH", "")
			OR StringAt((m_current + 1), 5, "HERES", ""))
		{
			MetaphAdd("T");
			AdvanceCounter(3, 2);
			return true;
		}
	}

	return false;
}

/**
 * Encode "-TCH-", reliably X ("sh", or in this case, "ch")
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_TCH()
{
	if(StringAt((m_current + 1), 2, "CH", ""))
	{
		MetaphAdd("X");
		m_current += 3;
		return true;
	}

	return false;
}

/**
 * Encode the many cases where americans are aware that a certain word is
 * french and know to not pronounce the 'T'
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_Silent_French_T()
{
	// french silent T familiar to americans
	if(((m_current == m_last) AND StringAt((m_current - 4), 5, "MONET", "GENET", "CHAUT", ""))
		OR StringAt((m_current - 2), 9, "POTPOURRI", "")
		OR StringAt((m_current - 3), 9, "BOATSWAIN", "")
		OR StringAt((m_current - 3), 8, "MORTGAGE", "")
		OR (StringAt((m_current - 4), 5, "BERET", "BIDET", "FILET", "DEBUT", "DEPOT", "PINOT", "TAROT", "")
		OR StringAt((m_current - 5), 6, "BALLET", "BUFFET", "CACHET", "CHALET", "ESPRIT", "RAGOUT", "GOULET",
										"CHABOT", "BENOIT", "")
		OR StringAt((m_current - 6), 7, "GOURMET", "BOUQUET", "CROCHET", "CROQUET", "PARFAIT", "PINCHOT",
										"CABARET", "PARQUET", "RAPPORT", "TOUCHET", "COURBET", "DIDEROT", "")
		OR StringAt((m_current - 7), 8, "ENTREPOT", "CABERNET", "DUBONNET", "MASSENET", "MUSCADET", "RICOCHET", "ESCARGOT", "")
		OR StringAt((m_current - 8), 9, "SOBRIQUET", "CABRIOLET", "CASSOULET", "OUBRIQUET", "CAMEMBERT", ""))
		AND !StringAt((m_current + 1), 2, "AN", "RY", "IC", "OM", "IN", ""))
	{
		m_current++;
		return true;
	}

	return false;
}

/**
 * Encode "-TU<N,L,A,O>-" in cases where it is pronounced
 * X ("sh", or in this case, "ch")
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_TUN_TUL_TUA_TUO()
{
	// e.g. "fortune", "fortunate"
	if(StringAt((m_current - 3), 6, "FORTUN", "")
		// e.g. "capitulate"
		OR (StringAt(m_current, 3, "TUL", "")
			AND (IsVowel(m_current - 1) AND IsVowel(m_current + 3)))
		// e.g. "obituary", "barbituate"
		OR  StringAt((m_current - 2), 5, "BITUA", "BITUE", "")
		// e.g. "actual"
		OR ((m_current > 1) AND StringAt(m_current, 3, "TUA", "TUO", "")))
	{
		MetaphAdd("X", "T");
		m_current++;
		return true;
	}

	return false;
}

/**
 * Encode "-T<vowel>-" forms where 'T' is pronounced as X
 * ("sh", or in this case "ch")
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_TUE_TEU_TEOU_TUL_TIE()
{
	// 'constituent', 'pasteur'
	if(StringAt((m_current + 1), 4, "UENT", "")
		OR StringAt((m_current - 4), 9, "RIGHTEOUS",  "")
		OR StringAt((m_current - 3), 7, "STATUTE",  "")
		OR StringAt((m_current - 3), 7, "AMATEUR",  "")
		// e.g. "blastula", "pasteur"
		OR (StringAt((m_current - 1), 5, "NTULE", "NTULA", "STULE", "STULA", "STEUR", ""))
		// e.g. "statue"
		OR (((m_current + 2) == m_last) AND StringAt(m_current, 3, "TUE", ""))
		// e.g. "constituency"
		OR StringAt(m_current, 5, "TUENC", "")
		// e.g. "statutory"
		OR StringAt((m_current - 3), 8, "STATUTOR", "")
		// e.g. "patience"
		OR (((m_current + 5) == m_last) AND StringAt(m_current, 6, "TIENCE", "")))
	{
		MetaphAdd("X", "T");
		AdvanceCounter(2, 1);
		return true;
	}

	return false;
}

/**
 * Encode "-TU-" forms in suffixes where it is usually
 * pronounced as X ("sh")
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_TUR_TIU_Suffixes()
{
	// 'adventure', 'musculature'
	if((m_current > 0) AND StringAt((m_current + 1), 3, "URE", "URA", "URI", "URY", "URO", "IUS", ""))
	{
		// exceptions e.g. 'tessitura', mostly from romance languages
		if ((StringAt((m_current + 1), 3, "URA", "URO", "")
				AND ((m_current + 3) == m_last))
				AND !StringAt((m_current - 3), 7, "VENTURA", "")
			// e.g. "kachaturian", "hematuria"
			OR StringAt((m_current + 1), 4, "URIA", ""))
		{
			MetaphAdd("T");
		}
		else
		{
			MetaphAdd("X", "T");
		}

		AdvanceCounter(2, 1);
		return true;
	}

	return false;
}

/**
 * Encode "-TI<O,A,U>-" as X ("sh"), except
 * in cases where it is part of a combining form,
 * or as J
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_TI()
{
	// '-tio-', '-tia-', '-tiu-'
	// except combining forms where T already pronounced e.g 'rooseveltian'
	if((StringAt((m_current + 1), 2, "IO", "") AND !StringAt((m_current - 1), 5, "ETIOL", ""))
		OR StringAt((m_current + 1), 3, "IAL", "")
		OR StringAt((m_current - 1), 5, "RTIUM", "ATIUM", "")
		OR ((StringAt((m_current + 1), 3, "IAN", "") AND (m_current > 0))
		AND !(StringAt((m_current - 4), 8, "FAUSTIAN", "")
		OR StringAt((m_current - 5), 9, "PROUSTIAN", "")
		OR StringAt((m_current - 2), 7, "TATIANA", "")
		OR(StringAt((m_current - 3), 7, "KANTIAN", "GENTIAN", "")
		OR StringAt((m_current - 8), 12, "ROOSEVELTIAN", "")))
		OR (((m_current + 2) == m_last)
		AND StringAt(m_current, 3, "TIA", "")
		// exceptions to above rules where the pronounciation is usually X
		AND !(StringAt((m_current - 3), 6, "HESTIA", "MASTIA", "")
		OR StringAt((m_current - 2), 5, "OSTIA", "")
		OR StringAt(0, 3, "TIA", "")
		OR StringAt((m_current - 5), 8, "IZVESTIA", "")))
		OR StringAt((m_current + 1), 4, "IATE", "IATI", "IABL", "IATO", "IARY", "")
		OR StringAt((m_current - 5), 9, "CHRISTIAN", "")))
	{
		if(((m_current == 2) AND StringAt(0, 4, "ANTI", ""))
			OR StringAt(0, 5, "PATIO", "PITIA", "DUTIA", ""))
		{
			MetaphAdd("T");
		}
		else if(StringAt((m_current - 4), 8, "EQUATION", ""))
		{
			MetaphAdd("J");
		}
		else
		{
			if(StringAt(m_current, 4, "TION", ""))
			{
				MetaphAdd("X");
			}
			else if(StringAt(0, 5, "KATIA", "LATIA", ""))
			{
				MetaphAdd("T", "X");
			}
			else
			{
				MetaphAdd("X", "T");
			}
		}

		AdvanceCounter(3, 1);
		return true;
	}

	return false;
}

/**
 * Encode "-TIENT-" where "TI" is pronounced X ("sh")
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_TIENT()
{
	// e.g. 'patient'
	if(StringAt((m_current + 1), 4, "IENT", ""))
	{
		MetaphAdd("X", "T");
		AdvanceCounter(3, 1);
		return true;
	}

	return false;
}

/**
 * Encode "-TSCH-" as X ("ch")
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_TSCH()
{
	//'deutsch'
	if(StringAt(m_current, 4, "TSCH", "")
		// combining forms in german where the 'T' is pronounced seperately
		AND !StringAt((m_current - 3), 4, "WELT", "KLAT", "FEST", ""))
	{
		// pronounced the same as "ch" in "chit" => X
		MetaphAdd("X");
		m_current += 4;
		return true;
	}

	return false;
}

/**
 * Encode "-TZSCH-" as X ("ch")
 *
 * "Neitzsche is peachy"
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_TZSCH()
{
	//'neitzsche'
	if(StringAt(m_current, 5, "TZSCH", ""))
	{
		MetaphAdd("X");
		m_current += 5;
		return true;
	}

	return false;
}

/**
 * Encodes cases where the 'H' in "-TH-" is the beginning of
 * another word in a combining form, special cases where it is
 * usually pronounced as 'T', and a special case where it has
 * become pronounced as X ("sh", in this case "ch")
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_TH_Pronounced_Separately()
{
	//'adulthood', 'bithead', 'apartheid'
	if(((m_current > 0)
			AND StringAt((m_current + 1), 4, "HOOD", "HEAD", "HEID", "HAND", "HILL", "HOLD",
											 "HAWK", "HEAP", "HERD", "HOLE", "HOOK", "HUNT",
											 "HUMO", "HAUS", "HOFF", "HARD", "")
			AND !StringAt((m_current - 3), 5, "SOUTH", "NORTH", ""))
		OR StringAt((m_current + 1), 5, "HOUSE", "HEART", "HASTE", "HYPNO", "HEQUE", "")
		// watch out for greek root "-thallic"
		OR (StringAt((m_current + 1), 4, "HALL", "")
			AND ((m_current + 4) == m_last)
			AND !StringAt((m_current - 3), 5, "SOUTH", "NORTH", ""))
		OR (StringAt((m_current + 1), 3, "HAM", "")
			AND ((m_current + 3) == m_last)
			AND !(StringAt(0, 6, "GOTHAM", "WITHAM", "LATHAM", "")
				 OR StringAt(0, 7, "BENTHAM", "WALTHAM", "WORTHAM", "")
				 OR StringAt(0, 8, "GRANTHAM", "")))
		OR (StringAt((m_current + 1), 5, "HATCH", "")
		AND !((m_current == 0) OR StringAt((m_current - 2), 8, "UNTHATCH", "")))
		OR StringAt((m_current - 3), 7, "WARTHOG", "")
		// and some special cases where "-TH-" is usually pronounced 'T'
		OR StringAt((m_current - 2), 6, "ESTHER", "")
		OR StringAt((m_current - 3), 6, "GOETHE", "")
		OR StringAt((m_current - 2), 8, "NATHALIE", ""))
	{
		// special case
		if (StringAt((m_current - 3), 7, "POSTHUM", ""))
		{
			MetaphAdd("X");
		}
		else
		{
			MetaphAdd("T");
		}
		m_current += 2;
		return true;
	}

	return false;
}

/**
 * Encode the "-TTH-" in "matthew", eating the redundant 'T'
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_TTH()
{
	// 'matthew' vs. 'outthink'
	if(StringAt(m_current, 3, "TTH", ""))
	{
		if (StringAt((m_current - 2), 5, "MATTH", ""))
		{
			MetaphAdd("0");
		}
		else
		{
			MetaphAdd("T0");
		}
		m_current += 3;
		return true;
	}

	return false;
}

/**
 * Encode "-TH-". 0 (zero) is used in Metaphone to encode this sound
 * when it is pronounced as a diphthong, either voiced or unvoiced
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_TH()
{
	if(StringAt(m_current, 2, "TH", "") )
	{
		//'-clothes-'
		if(StringAt((m_current - 3), 7, "CLOTHES", ""))
		{
			// vowel already encoded so skip right to S
			m_current += 3;
			return true;
		}

		//special case "thomas", "thames", "beethoven" or germanic words
		if(StringAt((m_current + 2), 4, "OMAS", "OMPS", "OMPK", "OMSO", "OMSE",
										"AMES", "OVEN", "OFEN", "ILDA", "ILDE", "")
			OR (StringAt(0, 4, "THOM", "")  AND (m_length == 4))
			OR (StringAt(0, 5, "THOMS", "")  AND (m_length == 5))
			OR StringAt(0, 4, "VAN ", "VON ", "")
			OR StringAt(0, 3, "SCH", ""))
		{
			MetaphAdd("T");

		}
		else
		{
			// give an 'etymological' 2nd
			// encoding for "smith"
			if(StringAt(0, 2, "SM", ""))
			{
				MetaphAdd("0", "T");
			}
			else
			{
				MetaphAdd("0");
			}
		}

		m_current += 2;
		return true;
	}

	return false;
}

/**
 * Encode "-V-"
 *
 */
void Metaphone3::Encode_V()
{
	// eat redundant 'V'
	if(m_inWord[m_current + 1] == 'V')
	{
		m_current += 2;
	}
	else
	{
		m_current++;
	}

	MetaphAddExactApprox("V", "F");
}

/**
 * Encode "-W-"
 *
 */
void Metaphone3::Encode_W()
{
	if(Encode_Silent_W_At_Beginning()
		OR Encode_WITZ_WICZ()
		OR Encode_WR()
		OR Encode_Initial_W_Vowel()
		OR Encode_WH()
		OR Encode_Eastern_European_W())
	{
		return;
	}

	// e.g. 'zimbabwe'
	if(m_encodeVowels
		AND StringAt(m_current, 2, "WE", "")
		AND ((m_current + 1) == m_last))
	{
		MetaphAdd("A");
	}

	//else skip it
	m_current++;

}

/**
 * Encode cases where 'W' is silent at beginning of word
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_Silent_W_At_Beginning()
{
	//skip these when at start of word
    if((m_current == 0)
		AND StringAt(m_current, 2, "WR", ""))
	{
        m_current += 1;
		return true;
	}

	return false;
}

/**
 * Encode polish patronymic suffix, mapping
 * alternate spellings to the same encoding,
 * and including easern european pronounciation
 * to the american so that both forms can
 * be found in a genealogy search
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_WITZ_WICZ()
{
	//polish e.g. 'filipowicz'
	if(((m_current + 3) == m_last) AND StringAt(m_current, 4, "WICZ", "WITZ", ""))
	{
		if(m_encodeVowels)
		{
			if((m_primary.length() > 0)
				AND m_primary[m_primary.length() - 1] == 'A')
			{
				MetaphAdd("TS", "FAX");
			}
			else
			{
				MetaphAdd("ATS", "FAX");
			}
		}
		else
		{
			MetaphAdd("TS", "FX");
		}
		m_current += 4;
		return true;
	}

	return false;
}

/**
 * Encode "-WR-" as R ('W' always effectively silent)
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_WR()
{
	//can also be in middle of word
	if(StringAt(m_current, 2, "WR", ""))
	{
		MetaphAdd("R");
		m_current += 2;
		return true;
	}

	return false;
}

/**
 * Encode "W-", adding central and eastern european
 * pronounciations so that both forms can be found
 * in a genealogy search
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_Initial_W_Vowel()
{
	if((m_current == 0) AND IsVowel(m_current + 1))
	{
		//Witter should match Vitter
		if(Germanic_Or_Slavic_Name_Beginning_With_W())
		{
			if(m_encodeVowels)
			{
				MetaphAddExactApprox("A", "VA", "A", "FA");
			}
			else
			{
				MetaphAddExactApprox("A", "V", "A", "F");
			}
		}
		else
		{
			MetaphAdd("A");
		}

		m_current++;
		// don't encode vowels twice
		m_current = SkipVowels(m_current);
		return true;
	}

	return false;
}

/**
 * Encode "-WH-" either as H, or close enough to 'U' to be
 * considered a vowel
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_WH()
{
	if(StringAt(m_current, 2, "WH", ""))
	{
		// cases where it is pronounced as H
		// e.g. 'who', 'whole'
		if((m_inWord[m_current + 2] == 'O')
			// exclude cases where it is pronounced like a vowel
			AND !(StringAt((m_current + 2), 4, "OOSH", "")
			OR StringAt((m_current + 2), 3, "OOP", "OMP", "ORL", "ORT", "")
			OR StringAt((m_current + 2), 2, "OA", "OP", "")))
		{
			MetaphAdd("H");
			AdvanceCounter(3, 2);
			return true;
		}
		else
		{
			// combining forms, e.g. 'hollowhearted', 'rawhide'
			if(StringAt((m_current + 2), 3, "IDE", "ARD", "EAD", "AWK", "ERD",
											"OOK", "AND", "OLE", "OOD", "")
				OR StringAt((m_current + 2), 4, "EART", "OUSE", "OUND", "")
				OR StringAt((m_current + 2), 5, "AMMER", ""))
			{
				MetaphAdd("H");
				m_current += 2;
				return true;
			}
			else if(m_current == 0)
			{
				MetaphAdd("A");
				m_current += 2;
				// don't encode vowels twice
				m_current = SkipVowels(m_current);
				return true;
			}
		}
		m_current += 2;
		return true;
	}

	return false;
}

/**
 * Encode "-W-" when in eastern european names, adding
 * the eastern european pronounciation to the american so
 * that both forms can be found in a genealogy search
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_Eastern_European_W()
{
	//Arnow should match Arnoff
	if(((m_current == m_last) AND IsVowel(m_current - 1))
		OR StringAt((m_current - 1), 5, "EWSKI", "EWSKY", "OWSKI", "OWSKY", "")
		OR (StringAt(m_current, 5, "WICKI", "WACKI", "") && ((m_current + 4) == m_last))
		OR StringAt(m_current, 4, "WIAK", "") AND ((m_current + 3) == m_last)
		OR StringAt(0, 3, "SCH", ""))
	{
		MetaphAddExactApprox("", "V", "", "F");
		m_current++;
		return true;
	}

	return false;
}

/**
 * Encode "-X-"
 *
 */
void Metaphone3::Encode_X()
{
	if(Encode_Initial_X()
		OR Encode_Greek_X()
		OR Encode_X_Special_Cases()
		OR Encode_X_To_H()
		OR Encode_X_Vowel()
		OR Encode_French_X_Final())
	{
		return;
	}

	// eat redundant 'X' or other redundant cases
	if(StringAt((m_current + 1), 1, "X", "Z", "S", "")
		// e.g. "excite", "exceed"
		OR StringAt((m_current + 1), 2, "CI", "CE", ""))
	{
		m_current += 2;
	}
	else
	{
		m_current++;
	}
}

/**
 * Encode initial X where it is usually pronounced as S
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_Initial_X()
{
	// current chinese pinyin spelling
	if(StringAt(0, 3, "XIA", "XIO", "XIE", "")
		OR StringAt(0, 2, "XU", ""))
	{
		MetaphAdd("X");
		m_current++;
		return true;
	}

	// else
	if((m_current == 0))
	{
		MetaphAdd("S");
		m_current++;
		return true;
	}

	return false;
}

/**
 * Encode X when from greek roots where it is usually pronounced as S
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_Greek_X()
{
	// 'xylophone', xylem', 'xanthoma', 'xeno-'
	if(StringAt((m_current + 1), 3, "YLO", "YLE", "ENO", "")
		OR StringAt((m_current + 1), 4, "ANTH", ""))
	{
		MetaphAdd("S");
		m_current++;
		return true;
	}

	return false;
}

/**
 * Encode special cases, "LUXUR-", "Texeira"
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_X_Special_Cases()
{
	// 'luxury'
	if(StringAt((m_current - 2), 5, "LUXUR", ""))
	{
		MetaphAddExactApprox("GJ", "KJ");
		m_current++;
		return true;
	}

	// 'texeira' portuguese/galician name
	if(StringAt(0, 7, "TEXEIRA", "")
		OR StringAt(0, 8, "TEIXEIRA", ""))
	{
		MetaphAdd("X");
		m_current++;
		return true;
	}

	return false;
}

/**
 * Encode special case spanish pronunciations of X
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_X_To_H()
{
	// TODO: look for other mexican indian words
	// where 'X' is usually pronounced this way
	if(StringAt((m_current - 2), 6, "OAXACA", "")
		OR StringAt((m_current - 3), 7, "QUIXOTE", ""))
	{
		MetaphAdd("H");
		m_current++;
		return true;
	}

	return false;
}

/**
 * Encode "-X-" in vowel contexts where it is usually
 * pronounced KX ("ksh")
 * account also for BBC pronounciation of => KS
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_X_Vowel()
{
	// e.g. "sexual", "connexion" (british), "noxious"
	if(StringAt((m_current + 1), 3, "UAL", "ION", "IOU", ""))
	{
		MetaphAdd("KX", "KS");
		AdvanceCounter(3, 1);
		return true;
	}

	return false;
}

/**
 * Encode cases of "-X", encoding as silent when part
 * of a french word where it is not pronounced
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_French_X_Final()
{
	//french e.g. "breaux", "paix"
	if(!((m_current == m_last)
		AND (StringAt((m_current - 3), 3, "IAU", "EAU", "IEU", "")
		OR StringAt((m_current - 2), 2, "AI", "AU", "OU", "OI", "EU", ""))) )
	{
		MetaphAdd("KS");
	}

	return false;
}

/**
 * Encode "-Z-"
 *
 */
void Metaphone3::Encode_Z()
{
	if(Encode_ZZ()
		OR Encode_ZU_ZIER_ZS()
		OR Encode_French_EZ()
		OR Encode_German_Z())
	{
		return;
	}

	if(Encode_ZH())
	{
		return;
	}
	else
	{
		MetaphAdd("S");
	}

	// eat redundant 'Z'
	if(m_inWord[m_current + 1] == 'Z')
	{
		m_current += 2;
	}
	else
	{
		m_current++;
	}
}

/**
 * Encode cases of "-ZZ-" where it is obviously part
 * of an italian word where "-ZZ-" is pronounced as TS
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_ZZ()
{
	// "abruzzi", 'pizza'
	if((m_inWord[m_current + 1] == 'Z')
		AND ((StringAt((m_current + 2), 1, "I", "O", "A", "")
		AND ((m_current + 2) == m_last))
		OR StringAt((m_current - 2), 9, "MOZZARELL", "PIZZICATO", "PUZZONLAN", "")))
	{
		MetaphAdd("TS", "S");
		m_current += 2;
		return true;
	}

	return false;
}

/**
 * Encode special cases where "-Z-" is pronounced as J
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_ZU_ZIER_ZS()
{
	if(((m_current == 1) AND StringAt((m_current - 1), 4, "AZUR", ""))
		OR (StringAt(m_current, 4, "ZIER", "")
			AND !StringAt((m_current - 2), 6, "VIZIER", "ROZIER", ""))
		OR StringAt(m_current, 3, "ZSA", ""))
	{
		MetaphAdd("J", "S");

		if(StringAt(m_current, 3, "ZSA", ""))
		{
			m_current += 2;
		}
		else
		{
			m_current++;
		}
		return true;
	}

	return false;
}

/**
 * Encode cases where americans recognize "-EZ" as part
 * of a french word where Z not pronounced
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_French_EZ()
{
	if(((m_current == 3) AND StringAt((m_current - 3), 4, "CHEZ", ""))
		OR StringAt((m_current - 5), 6, "RENDEZ", ""))
	{
		m_current++;
		return true;
	}

	return false;
}

/**
 * Encode cases where "-Z-" is in a german word
 * where Z => TS in german
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_German_Z()
{
	if(((m_current == 2) AND ((m_current + 1) == m_last) AND StringAt((m_current - 2), 4, "NAZI", ""))
		OR StringAt((m_current - 2), 6, "NAZIFY", "MOZART", "")
		OR StringAt((m_current - 3), 4, "HOLZ", "HERZ", "MERZ", "FITZ", "")
		OR (StringAt((m_current - 3), 4, "GANZ", "") AND !IsVowel(m_current + 1))
		OR StringAt((m_current - 4), 5, "STOLZ", "PRINZ", "")
		OR StringAt((m_current - 4), 7, "VENEZIA", "")
		OR StringAt((m_current - 3), 6, "HERZOG", "")
		// german words containing "sch-" but not schlimazel, schmooze
		OR ((m_inWord.find("SCH") != string::npos) && !(StringAt((m_last - 2), 3, "IZE", "OZE", "ZEL", "")))
		OR ((m_current > 0) AND StringAt(m_current, 4, "ZEIT", ""))
		OR StringAt((m_current - 3), 4, "WEIZ", ""))
	{
		if((m_current > 0) AND m_inWord[m_current - 1] == 'T')
		{
			MetaphAdd("S");
		}
		else
		{
			MetaphAdd("TS");
		}
		m_current++;
		return true;
	}

	return false;
}

/**
 * Encode "-ZH-" as J
 *
 * @return true if encoding handled in this routine, false if not
 *
 */
bool Metaphone3::Encode_ZH()
{
	//chinese pinyin e.g. 'zhao', also english "phonetic spelling"
	if(m_inWord[m_current + 1] == 'H')
	{
		MetaphAdd("J");
		m_current += 2;
		return true;
	}

	return false;
}

/**
 * Test for names derived from the swedish,
 * dutch, or slavic that should get an alternate
 * pronunciation of 'SV' to match the native
 * version
 *
 * @return true if swedish, dutch, or slavic derived name
 */
bool Metaphone3::Names_Beginning_With_SW_That_Get_Alt_SV()
{
	if(StringAt(0, 7, "SWANSON", "SWENSON", "SWINSON", "SWENSEN",
					  "SWOBODA", "")
		|| StringAt(0, 9, "SWIDERSKI", "SWARTHOUT", "")
		|| StringAt(0, 10, "SWEARENGIN", ""))
	{
		return true;
	}

	return false;
}

/**
 * Test for names derived from the german
 * that should get an alternate pronunciation
 * of 'XV' to match the german version spelled
 * "schw-"
 *
 * @return true if german derived name
 */
bool Metaphone3::Names_Beginning_With_SW_That_Get_Alt_XV()
{
	if(StringAt(0, 5, "SWART", "")
		|| StringAt(0, 6, "SWARTZ", "SWARTS", "SWIGER", "")
		|| StringAt(0, 7, "SWITZER", "SWANGER", "SWIGERT",
					      "SWIGART", "SWIHART", "")
		|| StringAt(0, 8, "SWEITZER", "SWATZELL", "SWINDLER", "")
		|| StringAt(0, 9, "SWINEHART", "")
		|| StringAt(0, 10, "SWEARINGEN", ""))
	{
		return true;
	}

	return false;
}

/**
 * Test whether or not the word in question
 * is a name of germanic or slavic origin, for
 * the purpose of determining whether to add an
 * alternate encoding of 'V'
 *
 * @return true if germanic or slavic name
 */
bool Metaphone3::Germanic_Or_Slavic_Name_Beginning_With_W()
{
	if(StringAt(0, 3, "WEE", "WIX", "WAX", "")
		OR StringAt(0, 4, "WOLF", "WEIS", "WAHL", "WALZ", "WEIL", "WERT",
						  "WINE", "WILK", "WALT", "WOLL", "WADA", "WULF",
						  "WEHR", "WURM", "WYSE", "WENZ", "WIRT", "WOLK",
						  "WEIN", "WYSS", "WASS", "WANN", "WINT", "WINK",
						  "WILE", "WIKE", "WIER", "WELK", "WISE", "")
		OR StringAt(0, 5, "WIRTH", "WIESE", "WITTE", "WENTZ", "WOLFF", "WENDT",
						  "WERTZ", "WILKE", "WALTZ", "WEISE", "WOOLF", "WERTH",
						  "WEESE", "WURTH", "WINES", "WARGO", "WIMER", "WISER",
						  "WAGER", "WILLE", "WILDS", "WAGAR", "WERTS", "WITTY",
						  "WIENS", "WIEBE", "WIRTZ", "WYMER", "WULFF", "WIBLE",
						  "WINER", "WIEST", "WALKO", "WALLA", "WEBRE", "WEYER",
						  "WYBLE", "WOMAC", "WILTZ", "WURST", "WOLAK", "WELKE",
						  "WEDEL", "WEIST", "WYGAN", "WUEST", "WEISZ", "WALCK",
						  "WEITZ", "WYDRA", "WANDA", "WILMA", "WEBER", "")
		OR StringAt(0, 6, "WETZEL", "WEINER", "WENZEL", "WESTER", "WALLEN", "WENGER",
						  "WALLIN", "WEILER", "WIMMER", "WEIMER", "WYRICK", "WEGNER",
						  "WINNER", "WESSEL", "WILKIE", "WEIGEL", "WOJCIK", "WENDEL",
						  "WITTER", "WIENER", "WEISER", "WEXLER", "WACKER", "WISNER",
						  "WITMER", "WINKLE", "WELTER", "WIDMER", "WITTEN", "WINDLE",
						  "WASHER", "WOLTER", "WILKEY", "WIDNER", "WARMAN", "WEYANT",
						  "WEIBEL", "WANNER", "WILKEN", "WILTSE", "WARNKE", "WALSER",
						  "WEIKEL", "WESNER", "WITZEL", "WROBEL", "WAGNON", "WINANS",
						  "WENNER", "WOLKEN", "WILNER", "WYSONG", "WYCOFF", "WUNDER",
						  "WINKEL", "WIDMAN", "WELSCH", "WEHNER", "WEIGLE", "WETTER",
						  "WUNSCH", "WHITTY", "WAXMAN", "WILKER", "WILHAM", "WITTIG",
						  "WITMAN", "WESTRA", "WEHRLE", "WASSER", "WILLER", "WEGMAN",
						  "WARFEL", "WYNTER", "WERNER", "WAGNER", "WISSER", "")
		OR StringAt(0, 7, "WISEMAN", "WINKLER", "WILHELM", "WELLMAN", "WAMPLER", "WACHTER",
						  "WALTHER", "WYCKOFF", "WEIDNER", "WOZNIAK", "WEILAND", "WILFONG",
						  "WIEGAND", "WILCHER", "WIELAND", "WILDMAN", "WALDMAN", "WORTMAN",
						  "WYSOCKI", "WEIDMAN", "WITTMAN", "WIDENER", "WOLFSON", "WENDELL",
						  "WEITZEL", "WILLMAN", "WALDRUP", "WALTMAN", "WALCZAK", "WEIGAND",
						  "WESSELS", "WIDEMAN", "WOLTERS", "WIREMAN", "WILHOIT", "WEGENER",
						  "WOTRING", "WINGERT", "WIESNER", "WAYMIRE", "WHETZEL", "WENTZEL",
						  "WINEGAR", "WESTMAN", "WYNKOOP", "WALLICK", "WURSTER", "WINBUSH",
						  "WILBERT", "WALLACH", "WYNKOOP", "WALLICK", "WURSTER", "WINBUSH",
						  "WILBERT", "WALLACH", "WEISSER", "WEISNER", "WINDERS", "WILLMON",
						  "WILLEMS", "WIERSMA", "WACHTEL", "WARNICK", "WEIDLER", "WALTRIP",
						  "WHETSEL", "WHELESS", "WELCHER", "WALBORN", "WILLSEY", "WEINMAN",
						  "WAGAMAN", "WOMMACK", "WINGLER", "WINKLES", "WIEDMAN", "WHITNER",
						  "WOLFRAM", "WARLICK", "WEEDMAN", "WHISMAN", "WINLAND", "WEESNER",
						  "WARTHEN", "WETZLER", "WENDLER", "WALLNER", "WOLBERT", "WITTMER",
						  "WISHART", "WILLIAM", "")
		OR StringAt(0, 8, "WESTPHAL", "WICKLUND", "WEISSMAN", "WESTLUND", "WOLFGANG", "WILLHITE",
						  "WEISBERG", "WALRAVEN", "WOLFGRAM", "WILHOITE", "WECHSLER", "WENDLING",
						  "WESTBERG", "WENDLAND", "WININGER", "WHISNANT", "WESTRICK", "WESTLING",
						  "WESTBURY", "WEITZMAN", "WEHMEYER", "WEINMANN", "WISNESKI", "WHELCHEL",
						  "WEISHAAR", "WAGGENER", "WALDROUP", "WESTHOFF", "WIEDEMAN", "WASINGER",
						  "WINBORNE", "")
		OR StringAt(0, 9, "WHISENANT", "WEINSTEIN", "WESTERMAN", "WASSERMAN", "WITKOWSKI", "WEINTRAUB",
					      "WINKELMAN", "WINKFIELD", "WANAMAKER", "WIECZOREK", "WIECHMANN", "WOJTOWICZ",
					      "WALKOWIAK", "WEINSTOCK", "WILLEFORD", "WARKENTIN", "WEISINGER", "WINKLEMAN",
						  "WILHEMINA", "")
		OR StringAt(0, 10, "WISNIEWSKI", "WUNDERLICH", "WHISENHUNT", "WEINBERGER", "WROBLEWSKI",
						   "WAGUESPACK", "WEISGERBER", "WESTERVELT", "WESTERLUND", "WASILEWSKI",
						   "WILDERMUTH", "WESTENDORF", "WESOLOWSKI", "WEINGARTEN", "WINEBARGER",
						   "WESTERBERG", "WANNAMAKER", "WEISSINGER", "")
		OR StringAt(0, 11, "WALDSCHMIDT", "WEINGARTNER", "WINEBRENNER", "")
		OR StringAt(0, 12, "WOLFENBARGER", "")
		OR StringAt(0, 13, "WOJCIECHOWSKI", ""))
	{
		return true;
	}
		return false;
}

/**
 * Test whether the word in question
 * is a name starting with 'J' that should
 * match names starting with a 'Y' sound.
 * All forms of 'John', 'Jane', etc, get
 * and alt to match e.g. 'Ian', 'Yana'. Joelle
 * should match 'Yael', 'Joseph' should match
 * 'Yusef'. German and slavic last names are
 * also included.
 *
 * @return true if name starting with 'J' that
 * should get an alternate encoding as a vowel
 */
bool Metaphone3::Names_Beginning_With_J_That_Get_Alt_Y()
{
	if(StringAt(0, 3, "JAN", "JON", "JAN", "JIN", "JEN", "")
		OR StringAt(0, 4, "JUHL", "JULY", "JOEL", "JOHN", "JOSH",
						  "JUDE", "JUNE", "JONI", "JULI", "JENA",
						  "JUNG", "JINA", "JANA", "JENI", "JOEL",
						  "JANN", "JONA", "JENE", "JULE", "JANI",
						  "JONG", "JOHN", "JEAN", "JUNG", "JONE",
						  "JARA", "JUST", "JOST", "JAHN", "JACO",
						  "JANG", "JUDE", "JONE", "")
		OR StringAt(0, 5, "JOANN", "JANEY", "JANAE", "JOANA", "JUTTA",
						  "JULEE", "JANAY", "JANEE", "JETTA", "JOHNA",
						  "JOANE", "JAYNA", "JANES", "JONAS", "JONIE",
						  "JUSTA", "JUNIE", "JUNKO", "JENAE", "JULIO",
						  "JINNY", "JOHNS", "JACOB", "JETER", "JAFFE",
						  "JESKE", "JANKE", "JAGER", "JANIK", "JANDA",
						  "JOSHI", "JULES", "JANTZ", "JEANS", "JUDAH",
						  "JANUS", "JENNY", "JENEE", "JONAH", "JONAS",
						  "JACOB", "JOSUE", "JOSEF", "JULES", "JULIE",
						  "JULIA", "JANIE", "JANIS", "JENNA", "JANNA",
						  "JEANA", "JENNI", "JEANE", "JONNA", "")
		OR StringAt(0, 6, "JORDAN", "JORDON", "JOSEPH", "JOSHUA", "JOSIAH",
						  "JOSPEH", "JUDSON", "JULIAN", "JULIUS", "JUNIOR",
						  "JUDITH", "JOESPH", "JOHNIE", "JOANNE", "JEANNE",
						  "JOANNA", "JOSEFA", "JULIET", "JANNIE", "JANELL",
						  "JASMIN", "JANINE", "JOHNNY", "JEANIE", "JEANNA",
						  "JOHNNA", "JOELLE", "JOVITA", "JOSEPH", "JONNIE",
						  "JANEEN", "JANINA", "JOANIE", "JAZMIN", "JOHNIE",
						  "JANENE", "JOHNNY", "JONELL", "JENELL", "JANETT",
						  "JANETH", "JENINE", "JOELLA", "JOEANN", "JULIAN",
						  "JOHANA", "JENICE", "JANNET", "JANISE", "JULENE",
						  "JOSHUA", "JANEAN", "JAIMEE", "JOETTE", "JANYCE",
						  "JENEVA", "JORDAN", "JACOBS", "JENSEN", "JOSEPH",
						  "JANSEN", "JORDON", "JULIAN", "JAEGER", "JACOBY",
						  "JENSON", "JARMAN", "JOSLIN", "JESSEN", "JAHNKE",
						  "JACOBO", "JULIEN", "JOSHUA", "JEPSON", "JULIUS",
						  "JANSON", "JACOBI", "JUDSON", "JARBOE", "JOHSON",
						  "JANZEN", "JETTON", "JUNKER", "JONSON", "JAROSZ",
						  "JENNER", "JAGGER", "JASMIN", "JEPSEN", "JORDEN",
						  "JANNEY", "JUHASZ", "JERGEN", "JAKOB", "")
		OR StringAt(0, 7, "JOHNSON", "JOHNNIE", "JASMINE", "JEANNIE", "JOHANNA",
						  "JANELLE", "JANETTE", "JULIANA", "JUSTINA", "JOSETTE",
						  "JOELLEN", "JENELLE", "JULIETA", "JULIANN", "JULISSA",
						  "JENETTE", "JANETTA", "JOSELYN", "JONELLE", "JESENIA",
						  "JANESSA", "JAZMINE", "JEANENE", "JOANNIE", "JADWIGA",
						  "JOLANDA", "JULIANE", "JANUARY", "JEANICE", "JANELLA",
						  "JEANETT", "JENNINE", "JOHANNE", "JOHNSIE", "JANIECE",
						  "JOHNSON", "JENNELL", "JAMISON", "JANSSEN", "JOHNSEN",
						  "JARDINE", "JAGGERS", "JURGENS", "JOURDAN", "JULIANO",
						  "JOSEPHS", "JHONSON", "JOZWIAK", "JANICKI", "JELINEK",
						  "JANSSON", "JOACHIM", "JANELLE", "JACOBUS", "JENNING",
						  "JANTZEN", "JOHNNIE",  "")
		OR StringAt(0, 8, "JOSEFINA", "JEANNINE", "JULIANNE", "JULIANNA", "JONATHAN",
						  "JONATHON", "JEANETTE", "JANNETTE", "JEANETTA", "JOHNETTA",
						  "JENNEFER", "JULIENNE", "JOSPHINE", "JEANELLE", "JOHNETTE",
						  "JULIEANN", "JOSEFINE", "JULIETTA", "JOHNSTON", "JACOBSON",
						  "JACOBSEN", "JOHANSEN", "JOHANSON", "JAWORSKI", "JENNETTE",
						  "JELLISON", "JOHANNES", "JASINSKI", "JUERGENS", "JARNAGIN",
						  "JEREMIAH", "JEPPESEN", "JARNIGAN", "JANOUSEK", "")
		OR StringAt(0, 9, "JOHNATHAN", "JOHNATHON", "JORGENSEN", "JEANMARIE", "JOSEPHINA",
					      "JEANNETTE", "JOSEPHINE", "JEANNETTA", "JORGENSON", "JANKOWSKI",
					      "JOHNSTONE", "JABLONSKI", "JOSEPHSON", "JOHANNSEN", "JURGENSEN",
					      "JIMMERSON", "JOHANSSON", "")
		OR StringAt(0, 10, "JAKUBOWSKI", ""))
		{
			return true;
		}
			return false;
}
