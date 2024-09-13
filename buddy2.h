#ifndef __BUDDY2_H__
#define __BUDDY2_H__
#include <stdint.h>

#define INVALID_BUDDY_FLAG      0xFFFFFFFF

#define LEFT_LEAF(index) ((index) * 2 + 1)
#define RIGHT_LEAF(index) ((index) * 2 + 2)
#define PARENT(index) ( ((index) + 1) / 2 - 1)

#define IS_POWER_OF_2(x) (!((x)&((x)-1)))
#define MAX(a, b) ((a) > (b) ? (a) : (b))

struct buddy_info_t {
    uint32_t offset:24;
    uint32_t level:8;
};

union buddy_info_union_t {
    struct buddy_info_t _info;
    uint32_t value;
};

struct buddy2 {
  uint32_t node_size;
  uint8_t level;           //level
  int8_t  max_level;
  uint8_t pad1[2];
  int8_t bits[0];
};

struct buddy2* buddy2_new( int size );
void buddy2_destroy( struct buddy2* self );

uint32_t buddy2_alloc(struct buddy2* self, int size);
void buddy2_free(struct buddy2* self, int offset);

int buddy2_size(struct buddy2* self, int offset);
void buddy2_dump(struct buddy2* self);

#endif//__BUDDY2_H__
