
// Common Valine/Weibo Emojis
// Source reference: https://unpkg.com/@waline/emojis@1.2.0/weibo/info.json
const WEIBO_EMOJI_BASE = 'https://unpkg.com/@waline/emojis@1.2.0/weibo/'
const WEIBO_PREFIX = 'weibo_'

// Known list from info.json to avoid 404s
const WEIBO_ITEMS = [
    "smile", "lovely", "blush", "grin", "laugh", "joy", "angry", "bye", "hammer", "kiss", "love", "mask", "money", "quiet", "rage", "sad", "shy", "sick", "slient", "smirk", "slap", "antic", "desire", "longing", "no_idea", "look_down", "clap", "yum", "sleep", "dizzy_face", "chuckle", "disappointed", "flushed", "heart_eyes", "no", "shuai", "suprised", "think", "vomit", "scream", "sleepy", "sob", "sunglasses", "greddy", "pick_nose", "annoyed", "awkward", "confused", "grievance", "poor", "wink", "rolling_eyes", "watermalon", "annoyed_left", "annoyed_right", "yawn", "hufen", "doge", "husky", "dog_annoyed", "dog_bye", "dog_consider", "dog_cry", "dog_joy", "dog_laugh", "dog_sweat", "dog_think", "dog_yum", "cat", "cat_annoyed", "cat_bye", "cat_cry", "cat_think", "girl_annoyed", "boy", "girl", "panda", "pig", "rabbit", "ultraman", "wool_group", "yan", "xi", "soap", "meng", "jiong", "geili", "shenma", "alpaca", "amorousness", "dog_sleepy", "dog_angry", "dog_kiss", "dog_love", "dog_sad", "dog_think", "dog_yum", "cat_sleepy", "cat_angry", "cat_kiss", "cat_love", "cat_sad", "cat_think", "cat_yum", "girl_sleepy", "girl_angry", "girl_kiss", "girl_love", "girl_sad", "girl_think", "girl_yum", "boy_sleepy", "boy_angry", "boy_kiss", "boy_love", "boy_sad", "boy_think", "boy_yum", "panda_sleepy", "panda_angry", "panda_kiss", "panda_love", "panda_sad", "panda_think", "panda_yum", "pig_sleepy", "pig_angry", "pig_kiss", "pig_love", "pig_sad", "pig_think", "pig_yum", "rabbit_sleepy", "rabbit_angry", "rabbit_kiss", "rabbit_love", "rabbit_sad", "rabbit_think", "rabbit_yum", "ultraman_sleepy", "ultraman_angry", "ultraman_kiss", "ultraman_love", "ultraman_sad", "ultraman_think", "ultraman_yum", "wool_group_sleepy", "wool_group_angry", "wool_group_kiss", "wool_group_love", "wool_group_sad", "wool_group_think", "wool_group_yum", "yan_sleepy", "yan_angry", "yan_kiss", "yan_love", "yan_sad", "yan_think", "yan_yum", "xi_sleepy", "xi_angry", "xi_kiss", "xi_love", "xi_sad", "xi_think", "xi_yum", "soap_sleepy", "soap_angry", "soap_kiss", "soap_love", "soap_sad", "soap_think", "soap_yum", "meng_sleepy", "meng_angry", "meng_kiss", "meng_love", "meng_sad", "meng_think", "meng_yum", "jiong_sleepy", "jiong_angry", "jiong_kiss", "jiong_love", "jiong_sad", "jiong_think", "jiong_yum", "geili_sleepy", "geili_angry", "geili_kiss", "geili_love", "geili_sad", "geili_think", "geili_yum", "shenma_sleepy", "shenma_angry", "shenma_kiss", "shenma_love", "shenma_sad", "shenma_think", "shenma_yum", "alpaca_sleepy", "alpaca_angry", "alpaca_kiss", "alpaca_love", "alpaca_sad", "alpaca_think", "alpaca_yum",
]

export const parseEmoji = (content: string): string => {
    if (!content) return ''

    // Replace :code: with <img> tag
    return content.replace(/:([a-zA-Z0-9_\s]+):/g, (match, rawCode) => {
        let code = rawCode.trim()

        // Special handling for "love you" which might be a legacy alias for "love"
        if (code === 'love you') code = 'love'

        // Check if it's in our known list
        if (WEIBO_ITEMS.includes(code)) {
            return `<img class="wl-emoji" src="${WEIBO_EMOJI_BASE}${WEIBO_PREFIX}${code}.png" alt="${code}" style="display:inline; height:1.2em; vertical-align:middle;" />`
        }

        return match // Return original if not found
    })
}
