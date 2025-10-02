#ifndef SNIPPETS_WRAPPER_H
#define SNIPPETS_WRAPPER_H

#include <stdint.h>
#include <stdbool.h>

#define MD5_LEN 16
// MAX_MULTIPLE_COMPONENTS will be defined by scan.h
// We need to undefine it first to avoid conflicts, then redefine with our value

typedef enum {
    WRAPPER_MATCH_NONE = 0,
    WRAPPER_MATCH_FILE = 1,
    WRAPPER_MATCH_SNIPPET = 2,
    WRAPPER_MATCH_BINARY = 3
} wrapper_match_t;

typedef struct wrapper_scan_input {
    uint8_t md5[MD5_LEN];
    char *file_path;
    uint32_t *hashes;
    uint32_t *lines;
    uint32_t hash_count;
    int total_lines;
} wrapper_scan_input_t;

typedef struct wrapper_match_info {
    char file_md5_hex[MD5_LEN * 2 + 1];
    int hits;
    int range_count;
    int *range_from;
    int *range_to;
    int *oss_line;
} wrapper_match_info_t;

typedef struct wrapper_scan_result {
    wrapper_match_t match_type;
    char *error_msg;
    wrapper_match_info_t *matches;
    int match_count;
} wrapper_scan_result_t;

wrapper_scan_result_t* snippets_wrapper_scan(wrapper_scan_input_t *input);
void snippets_wrapper_free_result(wrapper_scan_result_t *result);
void snippets_wrapper_init(const char *oss_db_name, bool enable_debug);
void snippets_wrapper_cleanup();

#endif