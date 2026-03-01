/**
 * A NewsDoc block representing a piece of content, metadata, or a link.
 *
 * All fields are optional because only non-empty values are included when a
 * document is converted to the scripting representation.
 */
interface Block {
  id?: string;
  uuid?: string;
  uri?: string;
  url?: string;
  type?: string;
  title?: string;
  rel?: string;
  role?: string;
  name?: string;
  value?: string;
  contenttype?: string;
  sensitivity?: string;
  /** Key-value string pairs. */
  data?: Record<string, string>;
  links?: Block[];
  content?: Block[];
  meta?: Block[];
}

/**
 * A NewsDoc document.
 *
 * All fields are optional because only non-empty values are included when a
 * document is converted to the scripting representation.
 */
interface Document {
  uuid?: string;
  type?: string;
  uri?: string;
  url?: string;
  title?: string;
  language?: string;
  content?: Block[];
  meta?: Block[];
  links?: Block[];
}

/**
 * Criteria object where all specified fields must match the corresponding block
 * fields. Only string-valued block fields can be matched.
 */
type BlockCriteria = Partial<
  Pick<
    Block,
    | "id"
    | "uuid"
    | "uri"
    | "url"
    | "type"
    | "title"
    | "rel"
    | "role"
    | "name"
    | "value"
    | "contenttype"
    | "sensitivity"
  >
>;

/** Predicate function that receives a block and returns whether it matches. */
type BlockPredicate = (block: Block) => boolean;

/** A matcher is either a criteria object or a predicate function. */
type Matcher = BlockCriteria | BlockPredicate;

/**
 * Block array parameter type. Accepts undefined so that optional document
 * fields like doc.meta can be passed directly — undefined is treated as an
 * empty array at runtime.
 */
type Blocks = Block[] | undefined;

/** NewsDoc block manipulation helpers. */
declare const nd: {
  /** Returns the first block matching the criteria, or null. */
  first_block(blocks: Blocks, matcher: Matcher): Block | null;

  /** Returns all blocks matching the criteria. */
  all_blocks(blocks: Blocks, matcher: Matcher): Block[];

  /** Returns true if any block matches the criteria. */
  has_block(blocks: Blocks, matcher: Matcher): boolean;

  /** Returns a new array with all matching blocks removed. */
  drop_blocks(blocks: Blocks, matcher: Matcher): Block[];

  /** Returns a new array keeping only the first matching block. */
  dedupe_blocks(blocks: Blocks, matcher: Matcher): Block[];

  /** Applies fn to all matching blocks, returning a new array. */
  alter_blocks(
    blocks: Blocks,
    matcher: Matcher,
    fn: (block: Block) => Block,
  ): Block[];

  /** Applies fn to only the first matching block, returning a new array. */
  alter_first_block(
    blocks: Blocks,
    matcher: Matcher,
    fn: (block: Block) => Block,
  ): Block[];

  /**
   * Finds the first match and applies fn to it. If no match is found, appends
   * defaultBlock instead.
   */
  upsert_block(
    blocks: Blocks,
    matcher: Matcher,
    defaultBlock: Block,
    fn: (block: Block) => Block,
  ): Block[];

  /** Replaces the first matching block with newBlock, or appends it. */
  add_or_replace_block(
    blocks: Blocks,
    matcher: Matcher,
    newBlock: Block,
  ): Block[];

  /**
   * Returns a value from a block's data map. Returns defaultValue (or an empty
   * string) if the key is missing.
   */
  get_data(block: Block, key: string, defaultValue?: string): string;

  /** Merges updates into data, returning a new map (no mutation). */
  upsert_data(
    data: Record<string, string>,
    updates: Record<string, string>,
  ): Record<string, string>;

  /**
   * Fills in missing or empty-string keys from defaults, returning a new map
   * (no mutation).
   */
  data_defaults(
    data: Record<string, string>,
    defaults: Record<string, string>,
  ): Record<string, string>;
};

/** HTML encoding, decoding, and sanitization helpers. */
declare const html: {
  /** HTML-encode a string (escape special characters). */
  encode(s: string): string;

  /** Decode HTML entities in a string. */
  decode(s: string): string;

  /** Strip all HTML tags using a strict policy. */
  strip(s: string): string;

  /** Strip HTML tags using a named bluemonday policy. */
  strip_policy(s: string, policyName: string): string;
};

/**
 * Entry point for the transformation script. Receives a document, transforms
 * it, and returns the result.
 *
 * The function name can be changed via WithFuncName on the Go side.
 */
declare function transform(doc: Document): Document;
