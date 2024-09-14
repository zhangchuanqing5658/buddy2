#include <stdio.h>
#include <stdlib.h>
#include <time.h>
#include "buddy2.h"

#define MAX_SIZE  64
#define MAX_MALLOC_SIZE MAX_SIZE/4
#define LOOP_COUNT 20

uint32_t check_bits[MAX_SIZE];
uint32_t malloc_array[MAX_SIZE];
uint32_t malloc_real[MAX_SIZE];
uint32_t malloc_tot;
uint32_t malloc_real_size;

void init_record_info() {
  for (int i = 0; i < MAX_SIZE; i++) {
    check_bits[i] = 0;
    malloc_array[i] = INVALID_BUDDY_FLAG;
    malloc_real[i]  = INVALID_BUDDY_FLAG;
    malloc_tot = 0;
    malloc_real_size = 0;
  }
}

//  long int random(void);
//  void srandom(unsigned int seed);

int main() {
  unsigned int seed = time(NULL);
  struct buddy2* buddy = buddy2_new(MAX_SIZE, MAX_MALLOC_SIZE);
  union buddy_info_union_t info;
  struct buddy2* buddy_replay = buddy2_new(MAX_SIZE, MAX_MALLOC_SIZE);
  
  srandom(seed);
  for (int cnt = 0; cnt < LOOP_COUNT; cnt++) {
    init_record_info();
    int index = 0;
    int malloc_failed = 0;
    while(1) {
      malloc_array[index] = random() % MAX_MALLOC_SIZE;
      info.value = buddy2_alloc(buddy, malloc_array[index]);
      if (info.value == INVALID_BUDDY_FLAG) {
        malloc_failed = malloc_array[index];
        malloc_array[index] = INVALID_BUDDY_FLAG;
        break;
      }

      //check vaild
      uint32_t bit_off = info._info.offset;
      for (int bit_index = 0; bit_index < (1<< info._info.level); bit_index++) {
        if (check_bits[bit_index + bit_off] != 0) {
          printf("malloc failed\n");
            goto replay_err;
        }
        check_bits[bit_index + bit_off] = 1;
      }

      //record result
      malloc_real[index] = info.value;
      malloc_tot  += malloc_array[index];
      malloc_real_size += 1 << info._info.level;
      index++;
    }

    printf("\n*****************\n");
    printf("loop:%d\n\tmalloc %d times: [%d - %d]; need:%d, max free:%d\n", cnt + 1, index, malloc_tot, malloc_real_size, malloc_failed, 1<<buddy->bits[0]);
    buddy2_dump(buddy);

    printf("\tmem bits:");
    for (int j = 0; j < MAX_SIZE; j++) {
      if (j % 4 == 0) printf(" ");
      printf("%d", check_bits[j]);
    }

    for (int j = 0; j < MAX_SIZE; j++) {
      if (malloc_array[j] == INVALID_BUDDY_FLAG) {
        break;
      }
      info.value = malloc_real[j];
      buddy2_free(buddy, info._info.offset);
    }

    if (buddy->bits[0] != buddy->level) {
      printf("free failed\n");
      goto replay_err;
    }

    printf("\n\n\tfree loop:%d\n", index);
    buddy2_dump(buddy);
    printf("\n*****************\n");
  }

  //error 
  buddy2_new(-1, 0);
  buddy2_alloc(buddy, 32);
  buddy2_destroy(buddy);
  buddy2_destroy(buddy_replay);
  return 0;

replay_err:
  printf("\n\n***********\n\terror malloc/free\n***********\n");

  printf("\n***********enter malloc***********\n");
  for (int i = 0; i < MAX_SIZE; i++) {
    if (malloc_array[i] == INVALID_BUDDY_FLAG) {
        break;
    }

    info.value = buddy2_alloc(buddy_replay, malloc_array[i]);

    printf("malloc[%d], buddy malloc[off:%d, size:%d]\n", malloc_array[i], info._info.offset, info._info.level);
    buddy2_dump(buddy_replay);
  }

  printf("\n***********enter free***********\n");
  for (int j = 0; j < MAX_SIZE; j++) {
      if (malloc_array[j] == INVALID_BUDDY_FLAG) {
        break;
      }
      info.value = malloc_real[j];
      buddy2_free(buddy_replay, info._info.offset);

      printf("free[%d], buddy free[off:%d, size:%d]\n", malloc_array[j], info._info.offset, info._info.level);
      buddy2_dump(buddy_replay);
    }
  return -1;
}
