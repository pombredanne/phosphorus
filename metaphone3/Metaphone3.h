#if !defined(METAPHONE3_H)
#define METAPHONE3_H
/**
 * 
 * Metaphone 3<br>
 * VERSION 2.1.3
 * by Lawrence Philips<br>
 * The Metaphone 3 algorithm may not be used in commercial applications
 * without express consent of Anthropomorphic Software LLC and Lawrence Philips.<br><br>
 * 
 * Metaphone 3 is designed to return an *approximate* phonetic key (and an alternate
 * approximate phonetic key when appropriate) that should be the same for English
 * words, and most names familiar in the United States, that are pronounced *similarly*.
 * The key value is *not* intended to be an *exact* phonetic, or even phonemic,
 * representation of the word. This is because a certain degree of 'fuzziness' has
 * proven to be useful in compensating for variations in pronunciation, as well as
 * mis-heard pronunciations. For example, although americans are not usually aware of it,
 * the letter 's' is normally pronounced 'z' at the end of words such as "sounds".<br><br>
 * 
 * The 'approximate' aspect of the encoding is implemented according to the following rules:<br><br>
 * 
 * (1) All vowels are encoded to the same value - 'A'. If the parameter encodeVowels
 * is set to false, only *initial* vowels will be encoded at all. If encodeVowels is set
 * to true, 'A' will be encoded at all places in the word that any vowels are normally
 * pronounced. 'W' as well as 'Y' are treated as vowels. Although there are differences in
 * the pronunciation of 'W' and 'Y' in different circumstances that lead to their being
 * classified as vowels under some circumstances and as consonants in others, for the purposes
 * of the 'fuzziness' component of the Soundex and Metaphone family of algorithms they will
 * be always be treated here as vowels.<br><br>
 *
 * (2) Voiced and un-voiced consonant pairs are mapped to the same encoded value. This
 * means that:<br>
 * 'D' and 'T' -> 'T'<br>
 * 'B' and 'P' -> 'P'<br>
 * 'G' and 'K' -> 'K'<br>
 * 'Z' and 'S' -> 'S'<br>
 * 'V' and 'F' -> 'F'<br><br>
 *
 * - In addition to the above voiced/unvoiced rules, 'CH' and 'SH' -> 'X', where 'X'
 * represents the "-SH-" and "-CH-" sounds in Metaphone 3 encoding.<br><br>
 *
 * - Also, the sound that is spelled as "TH" in English is encoded to '0' (zero symbol). (Although
 * Americans are not usually aware of it, "TH" is pronounced in a voiced (e.g. "that") as
 * well as an unvoiced (e.g. "theater") form, which are naturally mapped to the same encoding.)<br><br>
 * 
 * The encodings in this version of Metaphone 3 are according to pronunciations common in the
 * United States. This means that they will be inaccurate for consonant pronunciations that
 * are different in the United Kingdom, for example "tube" -> "CHOOBE" -> XAP rather than american TAP.<br><br>
 *
 * Metaphone 3 was preceded by by Soundex, patented in 1919, and Metaphone and Double Metaphone,
 * developed by Lawrence Philips. All of these algorithms resulted in a significant number of
 * incorrect encodings. Metaphone3 was tested against a database of about 100 thousand English words,
 * names common in the United States, and non-English words found in publications in the United States,
 * with an emphasis on words that are commonly mispronounced, prepared by the Moby Words website,
 * but with the Moby Words 'phonetic' encodings algorithmically mapped to Double Metaphone encodings.
 * Metaphone3 increases the accuracy of encoding of english words, common names, and non-English
 * words found in american publications from the 89% for Double Metaphone, to over 98%.<br><br>
 *
 * DISCLAIMER:
 * Anthropomorphic Software LLC claims only that Metaphone 3 will return correct encodings,
 * within the 'fuzzy' definition of correct as above, for a very high percentage of correctly
 * spelled English and commonly recognized non-English words. Anthropomorphic Software LLC
 * warns the user that a number of words remain incorrectly encoded, that misspellings may not
 * be encoded 'properly', and that people often have differing ideas about the pronunciation
 * of a word. Therefore, Metaphone 3 is not guaranteed to return correct results every time, and
 * so a desired target word may very well be missed. Creators of commercial products should
 * keep in mind that systems like Metaphone 3 produce a 'best guess' result, and should
 * condition the expectations of end users accordingly.<br><br>
 *
 * METAPHONE3 IS PROVIDED "AS IS" WITHOUT
 * WARRANTY OF ANY KIND. LAWRENCE PHILIPS AND ANTHROPOMORPHIC SOFTWARE LLC
 * MAKE NO WARRANTIES, EXPRESS OR IMPLIED, THAT IT IS FREE OF ERROR,
 * OR ARE CONSISTENT WITH ANY PARTICULAR STANDARD OF MERCHANTABILITY, 
 * OR THAT IT WILL MEET YOUR REQUIREMENTS FOR ANY PARTICULAR APPLICATION.
 * LAWRENCE PHILIPS AND ANTHROPOMORPHIC SOFTWARE LLC DISCLAIM ALL LIABILITY
 * FOR DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES RESULTING FROM USE 
 * OF THIS SOFTWARE. 
 *
 */


#include <string>
#include <algorithm>
#include <cctype>
#include <set>
using namespace std;

/** Default size of key storage allocation */
const int MAX_KEY_ALLOCATION = 32;
/** Default maximum length of encoded key. */
const int DEFAULT_MAX_KEY_LENGTH = 8;

 /**
 * Metaphone 3 is designed to return an <i>approximate</i> phonetic key (and an alternate
 * approximate phonetic key when appropriate) that should be the same for English
 * words, and most names familiar in the United States, that are pronounced "similarly".
 * The key value is <i>not</i> intended to be an exact phonetic, or even phonemic,
 * representation of the word. This is because a certain degree of 'fuzziness' has
 * proven to be useful in compensating for variations in pronunciation, as well as
 * mis-heard pronunciations. For example, although americans are not usually aware of it,
 * the letter 's' is normally pronounced 'z' at the end of words such as "sounds".<br><br>
 * 
 * The 'approximate' aspect of the encoding is implemented according to the following rules:<br><br>
 * 
 * (1) All vowels are encoded to the same value - 'A'. If the parameter encodeVowels
 * is set to false, only *initial* vowels will be encoded at all. If encodeVowels is set
 * to true, 'A' will be encoded at all places in the word that any vowels are normally
 * pronounced. 'W' as well as 'Y' are treated as vowels. Although there are differences in
 * the pronunciation of 'W' and 'Y' in different circumstances that lead to their being
 * classified as vowels under some circumstances and as consonants in others, for the purposes
 * of the 'fuzziness' component of the Soundex and Metaphone family of algorithms they will
 * be always be treated here as vowels.<br><br>
 *
 * (2) Voiced and un-voiced consonant pairs are mapped to the same encoded value. This
 * means that:<br>
 * 'D' and 'T' -> 'T'<br>
 * 'B' and 'P' -> 'P'<br>
 * 'G' and 'K' -> 'K'<br>
 * 'Z' and 'S' -> 'S'<br>
 * 'V' and 'F' -> 'F'<br><br>
 *
 * - In addition to the above voiced/unvoiced rules, 'CH' and 'SH' -> 'X', where 'X'
 * represents the "-SH-" and "-CH-" sounds in Metaphone 3 encoding.<br><br>
 *
 * - Also, the sound that is spelled as "TH" in English is encoded to '0' (zero symbol). (Although
 * americans are not usually aware of it, "TH" is pronounced in a voiced (e.g. "that") as
 * well as an unvoiced (e.g. "theater") form, which are naturally mapped to the same encoding.)<br><br>
 *
 * In the "Exact" encoding, voiced/unvoiced pairs are <i>not</i> mapped to the same encoding, except
 * for the voiced and unvoiced versions of 'TH', sounds such as 'CH' and 'SH', and for 'S' and 'Z',
 * so that the words whose metaph keys match will in fact be closer in pronunciation that with the
 * more approximate setting. Keep in mind that encoding settings for search strings should always
 * be exactly the same as the encoding settings of the stored metaph keys in your database!
 * Because of the considerably increased accuracy of Metaphone3, it is now possible to use this
 * setting and have a very good chance of getting a correct encoding.
 * <br><br>
 * In the Encode Vowels encoding, all non-initial vowels and diphthongs will be encoded to
 * 'A', and there will only be one such vowel encoding character between any two consonants.
 * It turns out that there are some surprising wrinkles to encoding non-initial vowels in
 * practice, pre-eminently in inversions between spelling and pronunciation such as e.g.
 * "wrinkle" => 'RANKAL', where the last two sounds are inverted when spelled.
 * <br><br>
 * The encodings in this version of Metaphone 3 are according to pronunciations common in the
 * United States. This means that they will be inaccurate for consonant pronunciations that
 * are different in the United Kingdom, for example "tube" -> "CHOOBE" -> XAP rather than american TAP.
 * <br><br>
 *
 */
class Metaphone3
{
		set<string> e_pron_set;

        /** Length of word sent in to be encoded, as 
		* measured at beginning of encoding. */
		int  m_length;

        /** Length of encoded key string. */
        unsigned short m_metaphLength;

        /** Flag whether or not to encode non-initial vowels. */
        bool m_encodeVowels;

        /** Flag whether or not to encode consonants as exactly 
		* as possible. */
		bool m_encodeExact;

        /** Internal copy of word to be encoded, allocated separately
		* from string pointed to in incoming parameter. */
        string m_inWord;

        /** Running copy of primary key. */
        string m_primary;

        /** Running copy of secondary key. */
        string m_secondary;

        /** Index of character in m_inWord currently being
		* encoded. */
	    int m_current;

        /** Index of last character in m_inWord. */
		int m_last;

		/** Flag that an AL inversion has already been done. */
		bool flag_AL_inversion;

		// Utility Functions

		void Init();
		int SkipVowels(int at);
		void AdvanceCounter(int ifNotEncodeVowels, int ifEncodeVowels);
		void ConvertExtendedASCIIChars();
		bool Front_Vowel(int at);
		bool SlavoGermanic();
		bool IsVowel(char inChar);
        bool IsVowel(int at);
        inline void MetaphAdd(const char* main);
        inline void MetaphAdd(const char* main, const char* alt);
        inline void MetaphAddExactApprox(const char* mainExact, const char* altExact, const char* main, const char* alt);
        inline void MetaphAddExactApprox(const char* mainExact, const char* main);
		// Multiplex String Comparator
        bool StringAt(int start, int length, ... );
		bool RootOrInflections(string inWord, string root);

		// Encoding Functions
		void Encode_Vowels();
		bool Skip_Silent_UE();
		bool E_Pronounced_At_End();
		bool O_Silent();
		bool E_Silent();
		bool Silent_Internal_E();
		bool E_Silent_Suffix(int at);
		bool E_Pronouncing_Suffix(int at);
		bool E_Pronounced_Exceptions();
		void Encode_E_Pronounced();

		void Encode_B();
		bool Encode_Silent_B();

		void Encode_C();
		bool Encode_Silent_C_At_Beginning();
		bool Encode_CA_To_S();
		bool Encode_CO_To_S();
		bool Encode_CH();
		bool Encode_CHAE();
		bool Encode_CH_To_H();
		bool Encode_Silent_CH();
		bool Encode_CH_To_X();
		bool Encode_English_CH_To_K();
		bool Encode_Germanic_CH_To_K();
		bool Encode_ARCH();
		bool Encode_Greek_CH_Initial();
		bool Encode_Greek_CH_Non_Initial();
		bool Encode_CCIA();
		bool Encode_CC();
		bool Encode_CK_CG_CQ();
		bool Encode_C_Front_Vowel();
		bool Encode_British_Silent_CE();
		bool Encode_CE();
		bool Encode_CI();
		bool Encode_Latinate_Suffixes();
		bool Encode_Silent_C();
		bool Encode_CZ();
		bool Encode_CS();

		void Encode_D();
		bool Encode_DG();
		bool Encode_DJ();
		bool Encode_DT_DD();
		bool Encode_D_To_J();
		bool Encode_DOUS();
		bool Encode_Silent_D();

		void Encode_F();

		void Encode_G();
		bool Encode_Silent_G_At_Beginning();
		bool Encode_GG();
		bool Encode_GK();

		bool Encode_GH_To_J();
		bool Encode_GH_To_H();
		bool Encode_UGHT();
		bool Encode_GH_H_Part_Of_Other_Word();
		bool Encode_GH_After_Consonant();
		bool Encode_Initial_GH();
		bool Encode_Silent_GH();
		bool Encode_GH_Special_Cases();
		bool Encode_GH_To_F();
		bool Encode_GH();

		bool Encode_Silent_G();
		bool Encode_GN();
		bool Encode_GL();
		bool Initial_G_Soft();
		bool Encode_Initial_G_Front_Vowel();
		bool Encode_NGER();
		bool Encode_GER();
		bool Encode_GEL();
		bool Internal_Hard_G_Other();
		bool Internal_Hard_G_Open_Syllable();
		bool Internal_Hard_GEN_GIN_GET_GIT();
		bool Internal_Hard_NG();
		bool Internal_Hard_G();
		bool Hard_GE_At_End();
		bool Encode_Non_Initial_G_Front_Vowel();
		bool Encode_GA_To_J();

		void Encode_H();
		bool Encode_Initial_Silent_H();
		bool Encode_Initial_HS();
		bool Encode_Initial_HU_HW();
		bool Encode_Non_Initial_Silent_H();
		bool Encode_H_Pronounced();

		void Encode_J();
		bool Encode_Spanish_J();
		bool Encode_German_J();
		bool Encode_Spanish_OJ_UJ();
		bool Encode_J_To_J();
		bool Encode_Spanish_J_2();
		bool Encode_J_As_Vowel();
		void Encode_Other_J();

		void Encode_K();
		bool Encode_Silent_K();

		void Encode_L();
		void Interpolate_Vowel_When_Cons_L_At_End();
		bool Encode_LELY_To_L();
		bool Encode_COLONEL();
		bool Encode_French_AULT();
		bool Encode_French_EUIL();
		bool Encode_French_OULX();
		bool Encode_Silent_L_In_LM();
		bool Encode_Silent_L_In_LK_LV();
		bool Encode_Silent_L_In_OULD();
		bool Encode_LL_As_Vowel_Special_Cases();
		bool Encode_LL_As_Vowel();
		bool Encode_LL_As_Vowel_Cases();
		bool Encode_Vowel_LE_Transposition(int save_current);
		bool Encode_Vowel_Preserve_Vowel_After_L( int save_current);
		void Encode_LE_Cases(int save_current);

		void Encode_M();
		bool Encode_Silent_M_At_Beginning();
		bool Encode_MR_And_MRS();
		bool Encode_MAC();
		bool Encode_MPT();
		bool Test_Silent_MB_1();
		bool Test_Pronounced_MB();
		bool Test_Silent_MB_2();
		bool Test_Pronounced_MB_2();
		bool Test_MN();
		void Encode_MB();

		void Encode_N();
		bool Encode_NCE();

		void Encode_P();
		bool Encode_Silent_P_At_Beginning();
		bool Encode_PT();
		bool Encode_PH();
		bool Encode_PPH();
		bool Encode_RPS();
		bool Encode_COUP();
		bool Encode_PNEUM();
		bool Encode_PSYCH();
		bool Encode_PSALM();
		void Encode_PB();

		void Encode_Q();

		void Encode_R();
		bool Encode_RZ();
		bool Test_Silent_R();
		bool Encode_Vowel_RE_Transposition();

		void Encode_S();
		bool Encode_Special_SW();
		bool Encode_SKJ();
		bool Encode_SJ();
		bool Encode_Silent_French_S_Final();
		bool Encode_Silent_French_S_Internal();
		bool Encode_ISL();
		bool Encode_STL();
		bool Encode_Christmas();
		bool Encode_STHM();
		bool Encode_ISTEN();
		bool Encode_Sugar();
		bool Encode_SH();
		bool Encode_SCH();
		bool Encode_SUR();
		bool Encode_SU();
		bool Encode_SSIO();
		bool Encode_SS();
		bool Encode_SIA();
		bool Encode_SIO();
		bool Encode_Anglicisations();
		bool Encode_SC();
		bool Encode_SEA_SUI_SIER();
		bool Encode_SEA();

		void Encode_T();
		bool Encode_T_Initial();
		bool Encode_TCH();
		bool Encode_Silent_French_T();
		bool Encode_TUN_TUL_TUA_TUO();
		bool Encode_TUE_TEU_TEOU_TUL_TIE();
		bool Encode_TUR_TIU_Suffixes();
		bool Encode_TI();
		bool Encode_TIENT();
		bool Encode_TSCH();
		bool Encode_TZSCH();
		bool Encode_TH_Pronounced_Separately();
		bool Encode_TTH();
		bool Encode_TH();

		void Encode_V();

		void Encode_W();
		bool Encode_Silent_W_At_Beginning();
		bool Encode_WITZ_WICZ();
		bool Encode_WR();
		bool Encode_Initial_W_Vowel();
		bool Encode_WH();
		bool Encode_Eastern_European_W();

		void Encode_X();
		bool Encode_Initial_X();
		bool Encode_Greek_X();
		bool Encode_X_Special_Cases();
		bool Encode_X_To_H();
		bool Encode_X_Vowel();
		bool Encode_French_X_Final();

		void Encode_Z();
		bool Encode_ZZ();
		bool Encode_ZU_ZIER_ZS();
		bool Encode_French_EZ();
		bool Encode_German_Z();
		bool Encode_ZH();

		bool Names_Beginning_With_SW_That_Get_Alt_SV();
		bool Names_Beginning_With_SW_That_Get_Alt_XV();
		bool Germanic_Or_Slavic_Name_Beginning_With_W();
		bool Names_Beginning_With_J_That_Get_Alt_Y();

public:
		
		/**
		 * Constructor, default. This constructor is most convenient when
		 * encoding more than one word at a time. The word to encode can
		 * be set using SetWord(char *). 
		 */
		Metaphone3();

		/**
		 * Constructor, parameterized. The Metaphone3 object will
		 * be initialized with the incoming string, and can be called
		 * on to encode this string. This constructor is most convenient
		 * when only one word needs to be encoded.
		 * 
		 * @param in pointer to char string of word to be encoded. 
		 */
        Metaphone3(const char* in);

		/**
		 * Sets word to be encoded.
		 * 
		 * @param in pointer to EXTERNALLY ALLOCATED char string of 
		 * the word to be encoded. 
		 */
        void SetWord(const char* in);

		/**
		 * Sets length allocated for output keys.
		 * If incoming number is greater than maximum allowable 
		 * length returned by GetMaximumKeyLength(), set key length
		 * to maximum key length and return false;  otherwise, set key 
		 * length to parameter value and return true.
		 * 
		 * @param inKeyLength new length of key.
		 * @return true if able to set key length to requested value. 
		 */
        bool SetKeyLength(unsigned short inKeyLength);

		/** Retrieves maximum number of characters currently allocated for encoded key. 
		 *
		 * @return short integer representing the length allowed for the key.
		 */
        unsigned short GetKeyLength(){return m_metaphLength;}

		/** Retrieves maximum number of characters allowed for encoded key. 
		 *
		 * @return short integer representing the length of allocated storage for the key.
		 */
        unsigned short GetMaximumKeyLength(){return (unsigned short)MAX_KEY_ALLOCATION;}

		/** Sets flag that causes Metaphone3 to encode non-initial vowels. However, even 
		 * if there are more than one vowel sound in a vowel sequence (i.e. 
		 * vowel diphthong, etc.), only one 'A' will be encoded before the next consonant or the
		 * end of the word.
		 *
		 * @param inEncodeVowels Non-initial vowels encoded if true, not if false. 
		 */
        void SetEncodeVowels(bool inEncodeVowels){m_encodeVowels = inEncodeVowels;}

		/** Retrieves setting determining whether or not non-initial vowels will be encoded. 
		 *
		 * @return true if the Metaphone3 object has been set to encode non-initial vowels, false if not.
		 */
        bool GetEncodeVowels(){return m_encodeVowels;}

		/** Sets flag that causes Metaphone3 to encode consonants as exactly as possible.
		 * This does not include 'S' vs. 'Z', since americans will pronounce 'S' at the
		 * at the end of many words as 'Z', nor does it include "CH" vs. "SH". It does cause
		 * a distinction to be made between 'B' and 'P', 'D' and 'T', 'G' and 'K', and 'V'
		 * and 'F'.
		 *
		 * @param inEncodeExact consonants to be encoded "exactly" if true, not if false. 
		 */
        void SetEncodeExact(bool inEncodeExact){m_encodeExact = inEncodeExact;}

		/** Retrieves setting determining whether or not consonants will be encoded "exactly".
		 *
		 * @return true if the Metaphone3 object has been set to encode "exactly", false if not.
		 */
        bool GetEncodeExact(){return m_encodeExact;}

		/** Retrives primary encoded key.
		 *
		 * @return a character pointer to the primary encoded key
		 */
		const char* GetMetaph(){return m_primary.c_str();}

		/** Retrives alternate encoded key, if any. 
		 *
		 * @return a character pointer to the alternate encoded key
		 */
		const char* GetAlternateMetaph(){return m_secondary.c_str();}

		/** Encodes input string to one or two key values according to Metaphone 3 rules.
		 *
		 * Uses internal allocations to store key results. 
		 * PLEASE NOTE, that in order to retrieve the output encoded key strings, the
		 * programmer must use GetMetaph() and GetAlternateMetaph(). 
		 */
		void Encode();

};
#endif // !defined(METAPHONE3_H)
