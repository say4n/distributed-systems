import java.io.BufferedReader;
import java.io.IOException;
import java.io.StringReader;

import org.apache.hadoop.conf.Configuration;
import org.apache.hadoop.fs.Path;
import org.apache.hadoop.io.DoubleWritable;
import org.apache.hadoop.io.Text;
import org.apache.hadoop.mapreduce.Job;
import org.apache.hadoop.mapreduce.Mapper;
import org.apache.hadoop.mapreduce.Reducer;
import org.apache.hadoop.mapreduce.lib.input.FileInputFormat;
import org.apache.hadoop.mapreduce.lib.output.FileOutputFormat;

public class MatmulPrimary {

  public static class PrimaryMapper
       extends Mapper<Object, Text, Text, DoubleWritable>{

    public void map(Object key, Text value, Context context
                    ) throws IOException, InterruptedException {
        BufferedReader br = new BufferedReader(new StringReader(value.toString()));

        String line = br.readLine();
        while (line != null && !line.isEmpty()) {
            String[] values = line.split(",");
            int numRepeat;
            String r, c;
            double val;

            numRepeat = Integer.parseInt(values[0]);
            r = values[1];
            c = values[2];

            boolean isFirstMatrix = Integer.parseInt(values[4]) == 1;

            val = Double.parseDouble(values[3]);

            for(int i = 0; i < numRepeat; i++) {
                Text ikey = new Text();

                if (isFirstMatrix) {
                    ikey.set(r + "," + c + "," + Integer.toString(i));
                } else {
                    ikey.set(Integer.toString(i) + "," + r + "," + c);
                }
                DoubleWritable ival = new DoubleWritable(val);
                context.write(ikey, ival);
            }
        }
    }
  }

  public static class PrimaryReducer
       extends Reducer<Text,DoubleWritable,Text,DoubleWritable> {
    private DoubleWritable result = new DoubleWritable();

    public void reduce(Text key, Iterable<DoubleWritable> values,
                       Context context
                       ) throws IOException, InterruptedException {
      double product = 1;
      for (DoubleWritable val : values) {
        product *= val.get();
      }
      result.set(product);
      context.write(key, result);
    }
  }

  public static void main(String[] args) throws Exception {
    Configuration conf = new Configuration();
    Job job = Job.getInstance(conf, "matrix multiplication");

    job.setJarByClass(MatmulPrimary.class);
    job.setMapperClass(PrimaryMapper.class);
    job.setCombinerClass(PrimaryReducer.class);
    job.setReducerClass(PrimaryReducer.class);
    job.setOutputKeyClass(Text.class);
    job.setOutputValueClass(DoubleWritable.class);

    FileInputFormat.addInputPath(job, new Path(args[0]));
    FileOutputFormat.setOutputPath(job, new Path(args[1]));

    System.exit(job.waitForCompletion(true) ? 0 : 1);
  }
}