import java.io.IOException;

import org.apache.hadoop.conf.Configuration;
import org.apache.hadoop.fs.Path;
import org.apache.hadoop.io.DoubleWritable;
import org.apache.hadoop.io.Text;
import org.apache.hadoop.mapreduce.Job;
import org.apache.hadoop.mapreduce.Mapper;
import org.apache.hadoop.mapreduce.Reducer;
import org.apache.hadoop.mapreduce.lib.input.FileInputFormat;
import org.apache.hadoop.mapreduce.lib.output.FileOutputFormat;

public class MatmulSecondary {

  public static class SecondaryMapper
       extends Mapper<Object, Text, Text, DoubleWritable>{

    public void map(Object key, Text value, Context context
                    ) throws IOException, InterruptedException {

        String[] valueString = value.toString().split("\t");
        String[] keyString = valueString[0].split(",");
        String i = keyString[0], k = keyString[1], j = keyString[2];

        Text ikey = new Text(i + "," + j);
        DoubleWritable ival = new DoubleWritable(Double.parseDouble(valueString[1]));

        context.write(ikey, ival);
    }
  }

  public static class SecondaryReducer
       extends Reducer<Text,DoubleWritable,Text,DoubleWritable> {
    private DoubleWritable result = new DoubleWritable();

    public void reduce(Text key, Iterable<DoubleWritable> values,
                       Context context
                       ) throws IOException, InterruptedException {
      double sum = 0;
      for (DoubleWritable val : values) {
        sum += val.get();
      }
      result.set(sum);

      context.write(key, result);
    }
  }

  public static void main(String[] args) throws Exception {
    Configuration conf = new Configuration();
    Job job = Job.getInstance(conf, "matrix multiplication");

    job.setJarByClass(MatmulSecondary.class);
    job.setMapperClass(SecondaryMapper.class);
    job.setCombinerClass(SecondaryReducer.class);
    job.setReducerClass(SecondaryReducer.class);
    job.setOutputKeyClass(Text.class);
    job.setOutputValueClass(DoubleWritable.class);

    FileInputFormat.addInputPath(job, new Path(args[0]));
    FileOutputFormat.setOutputPath(job, new Path(args[1]));

    System.exit(job.waitForCompletion(true) ? 0 : 1);
  }
}